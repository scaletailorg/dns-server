//go:build generate

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/AdguardTeam/AdGuardDNS/internal/agd"
	"github.com/AdguardTeam/AdGuardDNS/internal/agdhttp"
	"github.com/AdguardTeam/golibs/log"
)

func main() {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, countriesASNURL, nil)
	check(err)

	req.Header.Add("User-Agent", agdhttp.UserAgent())

	resp, err := c.Do(req)
	check(err)
	defer log.OnCloserError(resp.Body, log.ERROR)

	out, err := os.OpenFile("./asntops.go", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o664)
	check(err)
	defer log.OnCloserError(out, log.ERROR)

	countryTopASNs := map[agd.Country][]agd.ASN{}
	err = json.NewDecoder(resp.Body).Decode(&countryTopASNs)
	check(err)

	allTopASNs := map[agd.ASN]struct{}{}
	for _, asns := range countryTopASNs {
		for _, asn := range asns {
			if asn != 0 {
				allTopASNs[asn] = struct{}{}
			}
		}
	}

	type templateData struct {
		AllTopASNs     map[agd.ASN]struct{}
		CountryTopASNs map[agd.Country][]agd.ASN
	}

	tmplData := &templateData{
		AllTopASNs:     allTopASNs,
		CountryTopASNs: countryTopASNs,
	}

	tmpl, err := template.New("main").Parse(tmplStr)
	check(err)

	err = tmpl.Execute(out, tmplData)
	check(err)
}

// countriesASNURL is the default URL to get the per-country top ASN statistics
// from.
const countriesASNURL = `https://static.adtidy.org/dns/countries_asn.json`

// tmplStr is the template of the generated Go code.
const tmplStr = `// Code generated by go run ./asntops_generate.go; DO NOT EDIT.

package geoip

import "github.com/AdguardTeam/AdGuardDNS/internal/agd"

// allTopASNs contains all specially handled ASNs.
var allTopASNs = map[agd.ASN]struct{}{
{{- range $asn, $_ := .AllTopASNs }}
	{{ printf "%-7s {}," ( printf "%d:" $asn ) }}
{{- end }}
}

// countryTopASNs is a mapping of a country to their top ASNs.
var countryTopASNs = map[agd.Country]agd.ASN{
{{- range $ctry, $ASNs := .CountryTopASNs }}
{{- if gt (len $ASNs) 0 }}
	agd.Country{{ $ctry }}: {{ index $ASNs 0 }},
{{- else }}
{{- continue }}
{{- end }}
{{- end }}
}
`

// check is a simple error checker.
func check(err error) {
	if err != nil {
		panic(err)
	}
}