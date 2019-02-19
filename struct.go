package ribbit

type (
	SummaryItem struct {
		Product string `json:"product"`
		Seqn    string `json:"seqn"`
		Flags   string `json:"flags"`
	}

	RegionItem struct {
		Buildconfig   string `json:"build_config"`
		Buildid       string `json:"build_id"`
		Cdnconfig     string `json:"cdn_config"`
		Keyring       string `json:"keyring,omitempty"`
		Region        string `json:"region"`
		Versionsname  string `json:"version_name"`
		Productconfig string `json:"product_config,omitempty"`
	}

	CdnItem struct {
		Name        string   `json:"-"`
		Region      string   `json:"region"`
		Path        string   `json:"path"`
		Hosts       string   `json:"-"`
		HostsList   []string `json:"hosts_list"`
		Servers     string   `json:"-"`
		ServersList []string `json:"servers_lists"`
		ConfigPath  string   `json:"config_path"`
	}
)
