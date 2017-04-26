package main

import (
	"fmt"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"
	"log"
	"time"
)

type Conf struct {
	Address  string        `required:"true"`
	Token    string        `required:"true"`
	Interval time.Duration `default:"1h"`
	CRLs     []string      `required:"true"`
	TokenTTL time.Duration `envconfig:"TOKEN_TTL", default:"10m"`
}

func main() {
	config := Conf{}
	if err := envconfig.Process("crl_rotate", &config); err != nil {
		log.Fatal(err)
	}

	log.Println("connecting to vault")
	vc, err := vaultapi.NewClient(&vaultapi.Config{Address: config.Address})
	if err != nil {
		log.Fatalf("error connecting to vault: %v\n", err)
	}

	vc.SetToken(config.Token)
	client := vc.Logical()

	go func() {
		tc := vc.Auth().Token()

		sec, err := tc.LookupSelf()
		if err != nil {
			log.Fatalf("error looking up token: %v\n", err)
		}

		if sec.Data["renewable"].(bool) {
			renewal := time.Duration(config.TokenTTL.Seconds()*0.8) * time.Second
			for {
				_, err := tc.RenewSelf(int(config.TokenTTL.Seconds()))
				if err != nil {
					log.Printf("error renewing token: %v\n", err)
				} else {
					log.Println("token renewed")
				}
				time.Sleep(renewal)
			}
		} else {
			log.Println("token not renewable. wont bother")
			return
		}
	}()

	ticker := time.NewTicker(config.Interval)
	log.Println("sleeping until next renewal interval")
	for _ = range ticker.C {
		log.Println("renewal interval hit")
		for _, crl := range config.CRLs {
			sec, err := client.Read(fmt.Sprintf("%v/crl/rotate", crl))
			if err != nil {
				log.Fatalf("renewal of %v failed: %v\n", crl, err)
			}

			if sec.Data["success"].(bool) {
				log.Printf("renewal of %v crl: success", crl)
			} else {
				log.Printf("renewal of %v crl: unknown error", crl)
			}
		}
		log.Println("sleeping until next renewal interval")
	}
}
