module github.com/Cyb3r-Jak3/cloudflare-utils

go 1.19

require (
	github.com/Cyb3r-Jak3/common/v5 v5.1.0
	github.com/cloudflare/cloudflare-go v0.65.0
	github.com/sirupsen/logrus v1.9.0
	github.com/sourcegraph/conc v0.3.0
	github.com/urfave/cli/v2 v2.25.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
)

// Until https://github.com/cloudflare/cloudflare-go/pull/1264 is merged
replace github.com/cloudflare/cloudflare-go => github.com/Cyb3r-Jak3/cloudflare-go v0.42.1-0.20230419002028-c9b20db7de2d
