package main

import (
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/glerchundi/kube2nginx/pkg"
)

const (
	cliName        = "kube2nginx"
	cliDescription = "kube2nginx configures and reloads (if needed) nginx based on Kubernetes."
)

func AddConfigFlags(fs *flag.FlagSet, c *pkg.Config) {
}

func main() {
	// configuration
	cfg := pkg.NewConfig()

	// flags
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&cfg.IngressesData, "ingresses-data", cfg.IngressesData, "Ingresses data.")
	fs.StringVar(&cfg.KubeMasterURL, "kube-master-url", cfg.KubeMasterURL, "URL to reach kubernetes master.")
	fs.StringVar(&cfg.Namespace, "namespace", cfg.Namespace, "If present, the namespace scope.")
	fs.StringVar(&cfg.Selector, "selector", cfg.Selector, "Filter resources by a user-provided selector.")
	fs.DurationVar(&cfg.ResyncInterval, "resync-interval", cfg.ResyncInterval, "Resync with kubernetes master every user-defined interval.")
	fs.StringVar(&cfg.NginxSrc, "nginx-src", cfg.NginxSrc, "nginx.conf template file path.")
	fs.StringVar(&cfg.NginxDest, "nginx-dst", cfg.NginxDest, "nginx.conf destination file path.")
	fs.IntVar(&cfg.NginxDestUid, "nginx-dst-uid", cfg.NginxDestUid, "nginx.conf destination file uid.")
	fs.IntVar(&cfg.NginxDestGid, "nginx-dst-gid", cfg.NginxDestGid, "nginx.conf destination file gid.")
	fs.StringVar(&cfg.NginxDestMode, "nginx-dst-mode", cfg.NginxDestMode, "nginx.conf destination file mode.")
	fs.StringVar(&cfg.NginxCheckCmd, "nginx-check-cmd", cfg.NginxCheckCmd, "nginx check command.")
	fs.StringVar(&cfg.NginxReloadCmd, "nginx-reload-cmd", cfg.NginxReloadCmd, "nginx reload command.")
	fs.SetNormalizeFunc(
		func(f *flag.FlagSet, name string) flag.NormalizedName {
			if strings.Contains(name, "_") {
				return flag.NormalizedName(strings.Replace(name, "_", "-", -1))
			}
			return flag.NormalizedName(name)
		},
	)

	// parse
	fs.Parse(os.Args[1:])

	// set from env (if present)
	fs.VisitAll(func(f *flag.Flag) {
		if !f.Changed {
			key := strings.ToUpper(strings.Join(
			[]string{
				cliName,
				strings.Replace(f.Name, "-", "_", -1),
			},
			"_",
			))
			val := os.Getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
			}
		}
	})

	// and then, run!
	k2n := pkg.NewKubeToNginx(cfg)
	k2n.Run()
}