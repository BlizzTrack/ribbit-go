package ribbit

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
	"github.com/mitchellh/mapstructure"
)

//products/%s/versions

type RibbitClient struct {
	Region string
	Proxy  string
}

var client *RibbitClient
var d net.Dialer
var timeout time.Duration

func NewRibbitClient(reg string) *RibbitClient {
	var region string
	if len(reg) <= 0 {
		region = "us"
	} else {
		region = reg
	}

	client = &RibbitClient{region, ""}
	d = net.Dialer{Timeout: 5 * time.Second, DualStack: true}
	return client
}

func NewRibbitClientProxy(server string) *RibbitClient {
	client = &RibbitClient{"", server}
	d = net.Dialer{Timeout: 5 * time.Second, DualStack: true}
	return client
}

func SetTimeout(out time.Duration) {
	timeout = out
}

func (client *RibbitClient) Summary() ([]SummaryItem, string, string, error) {
	data, raw, err := client.process("summary")
	if err != nil {
		return nil, "", "", err
	}

	var result []SummaryItem
	mapstructure.Decode(parseFile(data), &result)

	return result, getSeqn(data), raw, nil
}

func (client *RibbitClient) CDNS(game string) ([]CdnItem, string, string, error) {
	data, raw, err := client.process(fmt.Sprintf("products/%s/cdns", game))
	if err != nil {
		return nil, "", "", err
	}

	var result []CdnItem
	mapstructure.Decode(parseFile(data), &result)

	for i := 0; i < len(result); i++ {
		result[i].HostsList = strings.Split(result[i].Hosts, " ")
		result[i].ServersList = strings.Split(result[i].Servers, " ")
		result[i].Region = result[i].Name
	}

	return result, getSeqn(data), raw, nil
}

func (client *RibbitClient) Versions(game string) ([]RegionItem, string, string, error) {
	data, raw, err := client.process(fmt.Sprintf("products/%s/versions", game))
	if err != nil {
		return nil, "", "", err
	}

	var result []RegionItem
	mapstructure.Decode(parseFile(data), &result)

	return result, getSeqn(data), raw, nil
}

func (client *RibbitClient) BGDL(game string) ([]RegionItem, string, string, error) {
	data, raw, err := client.process(fmt.Sprintf("products/%s/bgdl", game))
	if err != nil {
		return nil, "", "", err
	}

	var result []RegionItem
	mapstructure.Decode(parseFile(data), &result)

	return result, getSeqn(data), raw, nil
}

func (client *RibbitClient) process(call string) (string, string, error) {
	d.Timeout = timeout
	d.Deadline = time.Now().Add(timeout)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()

	server := fmt.Sprintf("%s.version.battle.net:1119", client.Region)
	if client.Proxy != "" {
		server = client.Proxy
	}

	ribbitClient, err := d.DialContext(ctx, "tcp", server)
	if err != nil {
		return "", "", err
	}
	defer ribbitClient.Close()

	err = ribbitClient.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", "", err
	}
	err = ribbitClient.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", "", err
	}
	err = ribbitClient.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", "", err
	}

	_, err = fmt.Fprintf(ribbitClient, fmt.Sprintf("v1/%s\r\n", call))
	if err != nil {
		return "", "", err
	}

	data, err := ioutil.ReadAll(ribbitClient)
	if err != nil {
		return "", "", err
	}

	content := string(data)
	r := strings.NewReader(content)
	env, err := enmime.ReadEnvelope(r)
	if err != nil {
		return "", "", err
	}

	if env.Root == nil || env.Root.FirstChild == nil {
		return "", "", errors.New("root or firstchild of root is empty")
	}

	return string(env.Root.FirstChild.Content), content, nil
}

func (item SummaryItem) Versions() ([]RegionItem, string, string, error) {
	return client.Versions(item.Product)
}

func (item SummaryItem) BGDL() ([]RegionItem, string, string, error) {
	return client.BGDL(item.Product)
}

func (item SummaryItem) CDNS() ([]CdnItem, string, string, error) {
	return client.CDNS(item.Product)
}

// parser for version files... This will have to be changed to handle arrays in later build
func parseFile(file string) []map[string]string {
	lines := strings.Split(file, "\n")
	keys := strings.Split(lines[0], `|`)
	keysList := make([]string, len(keys))

	for i := 0; i < len(keys); i++ {
		keyList := strings.Split(keys[i], `!`)

		keysList[i] = strings.ToLower(keyList[0])
	}

	var data []map[string]string
	for i := 1; i < len(lines); i++ {
		if len(strings.TrimSpace(lines[i])) > 0 {
			if !strings.HasPrefix(lines[i], "#") {
				local := make(map[string]string)

				lineData := strings.Split(lines[i], `|`)

				for x := 0; x < len(keysList); x++ {
					if len(lineData[x]) > 0 {
						local[keysList[x]] = lineData[x]
					}
				}

				data = append(data, local)
			}
		}
	}

	return data
}

func getSeqn(file string) string {
	lines := strings.Split(file, "\n")
	for i := 1; i < len(lines); i++ {
		if len(strings.TrimSpace(lines[i])) > 0 {
			if strings.Contains(lines[i], "## seqn") {
				line := strings.Replace(lines[i], " ", "", -1)
				items := strings.Split(line, "=")
				return items[len(items)-1]
			}
		}
	}

	return ""
}
