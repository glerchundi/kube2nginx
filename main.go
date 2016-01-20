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
	fs.StringVar(&cfg.NginxTemplatePath, "nginx-template-path", cfg.NginxTemplatePath, "nginx.conf template path.")
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