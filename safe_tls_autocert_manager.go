package util

import (
	"crypto/tls"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type SafeTLSAutoCertManager struct {
	_private *autocert.Manager
}

func NewSafeAutoCertManager(tls_email_address string, domain_name_without_www string, ssl_cache_dir string, whitelist_prefixes []string) *SafeTLSAutoCertManager {
	// build up the list of TLS whitelisted domains: "example.com", "www.example.com", "www2.example.com" etc.
	tls_whitelist_domains := []string{domain_name_without_www}
	for _, prefix := range whitelist_prefixes {
		tls_whitelist_domains = append(tls_whitelist_domains, prefix+"."+domain_name_without_www)
	}
	m := &autocert.Manager{ //nolint:exhaustruct // this is from the example code in the official Go documentation
		Cache:      autocert.DirCache(ssl_cache_dir),
		Prompt:     autocert.AcceptTOS,
		Email:      tls_email_address,
		HostPolicy: autocert.HostWhitelist(tls_whitelist_domains...), // autocert doesn't support wildcard or regex
	}
	return &SafeTLSAutoCertManager{_private: m}
}

// SecureTLSConfig creates a new secure TLS config suitable for net/http.Server servers,
// supporting HTTP/2 and the tls-alpn-01 ACME challenge type.
func (m *SafeTLSAutoCertManager) GetSecureTLSConfig() *tls.Config {
	return &tls.Config{ //nolint:exhaustruct // I'm using Valsorda's example config: https://blog.cloudflare.com/exposing-go-on-the-internet/
		GetCertificate: m._private.GetCertificate,
		NextProtos: []string{
			"h2", "http/1.1", // enable HTTP/2
			acme.ALPNProto, // enable tls-alpn ACME challenges
		},
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.X25519, // Go 1.8 only
			tls.CurveP256,
		},
		MinVersion: tls.VersionTLS13,
	}
}

// Test your web server with these free services:
// https://www.ssllabs.com/ssltest/
// https://observatory.mozilla.org/
// https://internet.nl/

// Mozilla Security Guidelines:
// https://infosec.mozilla.org/guidelines/
// https://infosec.mozilla.org/guidelines/web_security.html
//
