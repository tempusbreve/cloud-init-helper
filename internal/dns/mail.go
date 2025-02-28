package dns

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MailConfig struct {
	api API
}

func WithAPI(api API) func(*MailConfig) { return func(c *MailConfig) { c.api = api } }

func NewMailConfig(options ...func(*MailConfig)) *MailConfig {
	cfg := &MailConfig{}

	for _, fn := range options {
		fn(cfg)
	}

	return cfg
}

type UpdateMailRecordsParams struct {
	Domain     string
	MXHosts    map[string]int
	Postmaster string
	DKIM       string

	Destructive bool
}

func (c *MailConfig) UpdateAllMailRecords(ctx context.Context, options UpdateMailRecordsParams) error {
	for name, fn := range map[string]func(context.Context, UpdateMailRecordsParams) error{
		"MX Records":     c.UpdateMXRecords,
		"SPF Records":    c.UpdateSPFRecords,
		"DKIM Record":    c.UpdateDKIMRecord,
		"DMARC Record":   c.UpdateDMARCRecord,
		"MTS-STS Record": c.UpdateMTSSTSRecord,
	} {
		if err := fn(ctx, options); err != nil {
			return fmt.Errorf("updating mail records (%s): %w", name, err)
		}
	}
	return nil
}

func (c *MailConfig) UpdateMXRecords(ctx context.Context, options UpdateMailRecordsParams) error {
	if options.Destructive {
		existing, err := c.api.GetRecords(ctx, options.Domain, "MX")
		if err != nil {
			return err
		}

		for _, rec := range existing {
			if err = c.api.DeleteRecord(ctx, rec.ID()); err != nil {
				return fmt.Errorf("deleting record: %v: %w", rec.ID(), err)
			}
		}
	}

	for host, weight := range options.MXHosts {
		if err := c.api.CreateMXRecord(ctx, options.Domain, host, weight); err != nil {
			return err
		}
	}

	return nil
}

func (c *MailConfig) UpdateSPFRecords(ctx context.Context, options UpdateMailRecordsParams) error {
	if options.Destructive {
		existing, err := c.api.GetRecords(ctx, options.Domain, "TXT")
		if err != nil {
			return err
		}

		for _, rec := range existing {
			if strings.Contains(rec.Content(), "v=spf1") {
				if err = c.api.DeleteRecord(ctx, rec.ID()); err != nil {
					return fmt.Errorf("deleting record: %v: %w", rec.ID(), err)
				}
			}
		}
	}

	rec := NewRecord(options.Domain, RecordTypeTXT, "v=spf1 mx ~all")
	return c.api.CreateRecord(ctx, rec)
}

func (c *MailConfig) UpdateDKIMRecord(ctx context.Context, options UpdateMailRecordsParams) error {
	name := "default._domainkey." + options.Domain

	dkim, err := getDKIMRecord(options)
	if err != nil {
		return err
	}

	if options.Destructive {
		existing, err := c.api.GetRecords(ctx, name, "TXT")
		if err != nil {
			return err
		}

		for _, rec := range existing {
			if err = c.api.DeleteRecord(ctx, rec.ID()); err != nil {
				return fmt.Errorf("deleting record: %v: %w", rec.ID(), err)
			}
		}
	}

	rec := NewRecord(name, RecordTypeTXT, dkim)
	return c.api.CreateRecord(ctx, rec)
}

func (c *MailConfig) UpdateDMARCRecord(ctx context.Context, options UpdateMailRecordsParams) error {
	name := "_dmarc." + options.Domain
	content := "v=DMARC1; p=quarantine; ruf=" + options.Postmaster

	if options.Destructive {
		existing, err := c.api.GetRecords(ctx, name, "TXT")
		if err != nil {
			return err
		}

		for _, rec := range existing {
			if strings.Contains(rec.Content(), "v=DMARC1") {
				if err = c.api.DeleteRecord(ctx, rec.ID()); err != nil {
					return fmt.Errorf("deleting record: %v: %w", rec.ID(), err)
				}
			}
		}
	}

	rec := NewRecord(name, RecordTypeTXT, content)
	return c.api.CreateRecord(ctx, rec)
}

func (c *MailConfig) UpdateMTSSTSRecord(ctx context.Context, options UpdateMailRecordsParams) error {
	for name, value := range map[string]string{
		"_mta-sts." + options.Domain:   "v=STSv1; id=1",
		"_smtp._tls." + options.Domain: "v=TLSRPTv1; rua=mailto:" + options.Postmaster,
	} {
		if options.Destructive {
			existing, err := c.api.GetRecords(ctx, name, "TXT")
			if err != nil {
				return err
			}

			for _, rec := range existing {
				if err = c.api.DeleteRecord(ctx, rec.ID()); err != nil {
					return fmt.Errorf("deleting record: %v: %w", rec.ID(), err)
				}
			}
		}

		rec := NewRecord(name, RecordTypeTXT, value)
		if err := c.api.CreateRecord(ctx, rec); err != nil {
			return err
		}
	}

	return nil
}

func getDKIMRecord(opts UpdateMailRecordsParams) (string, error) {
	if opts.DKIM != "" {
		return opts.DKIM, nil
	}

	df, err := os.Open(filepath.Join("/var/lib/maddy/dkim_keys", opts.Domain+"_default.dns"))
	if err != nil {
		return "", fmt.Errorf("opening dkim key file: %w", err)
	}
	defer df.Close()

	buf, err := io.ReadAll(df)
	if err != nil {
		return "", fmt.Errorf("reading dkim key: %w", err)
	}

	return string(buf), nil
}
