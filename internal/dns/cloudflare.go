package dns

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudflare/cloudflare-go"
)

var ErrInvalidRecordID = errors.New("invalid or missing record id")

type CloudFlareDNS struct {
	token    string
	zoneName string
}

func WithCFToken(token string) func(*CloudFlareDNS) {
	return func(d *CloudFlareDNS) { d.token = token }
}

func WithCFZoneName(zoneName string) func(*CloudFlareDNS) {
	return func(d *CloudFlareDNS) { d.zoneName = zoneName }
}

func NewCloudFlareDNS(options ...func(*CloudFlareDNS)) *CloudFlareDNS {
	dns := &CloudFlareDNS{}

	for _, fn := range options {
		fn(dns)
	}

	return dns
}

func (a *CloudFlareDNS) GetRecords(ctx context.Context, recordName, recordType string) ([]Record, error) {
	api, err := cfAPI(a.token)
	if err != nil {
		return nil, err
	}

	id, err := cfZoneID(api, "", a.zoneName)
	if err != nil {
		return nil, err
	}

	records, err := getRecords(ctx, api, id, recordType, recordName, "")
	if err != nil {
		return nil, err
	}

	var res []Record
	for _, rec := range records {
		res = append(res, cfToRecord(&rec))
	}

	return res, nil
}

func (a *CloudFlareDNS) CreateMXRecord(ctx context.Context, mailDomain string, mxHost string, weight int) error {
	api, err := cfAPI(a.token)
	if err != nil {
		return err
	}

	id, err := cfZoneID(api, "", a.zoneName)
	if err != nil {
		return err
	}

	return createMXRecord(ctx, api, id, mailDomain, mxHost, weight)
}

func (a *CloudFlareDNS) GetRecord(ctx context.Context, id any) (Record, error) {
	api, err := cfAPI(a.token)
	if err != nil {
		return nil, err
	}

	zid, err := cfZoneID(api, "", a.zoneName)
	if err != nil {
		return nil, err
	}

	if rid, ok := (id).(string); ok {
		rec, err := getRecord(ctx, api, zid, rid)
		return cfToRecord(&rec), err
	}

	return nil, fmt.Errorf("%w: %q", ErrInvalidRecordID, id)
}

func (a *CloudFlareDNS) CreateRecord(ctx context.Context, rec Record) error {
	api, err := cfAPI(a.token)
	if err != nil {
		return err
	}

	id, err := cfZoneID(api, "", a.zoneName)
	if err != nil {
		return err
	}

	return createRecord(ctx, api, id, string(rec.Type()), rec.Name(), rec.Content())
}

func (a *CloudFlareDNS) UpdateRecord(ctx context.Context, rec Record) error {
	api, err := cfAPI(a.token)
	if err != nil {
		return err
	}

	id, err := cfZoneID(api, "", a.zoneName)
	if err != nil {
		return err
	}

	if rid, ok := (rec.ID()).(string); ok {
		return updateRecord(ctx, api, id, rid, string(rec.Type()), rec.Name(), rec.Content())
	}

	return fmt.Errorf("%w: %q", ErrInvalidRecordID, rec.ID())
}

func (a *CloudFlareDNS) DeleteRecord(ctx context.Context, id any) error {
	api, err := cfAPI(a.token)
	if err != nil {
		return err
	}

	zid, err := cfZoneID(api, "", a.zoneName)
	if err != nil {
		return err
	}

	if rid, ok := (id).(string); ok {
		return deleteRecord(ctx, api, zid, rid)
	}

	return fmt.Errorf("%w: %q", ErrInvalidRecordID, id)
}

func cfToRecord(r *cloudflare.DNSRecord) Record {
	if r == nil {
		return nil
	}

	return record{
		id:      r.ID,
		name:    r.Name,
		rtype:   r.Type,
		content: r.Content,
		ttl:     r.TTL,
	}
}

func cfAPI(tok string) (*cloudflare.API, error) {
	return cloudflare.NewWithAPIToken(
		tok,
		cloudflare.UserAgent("cloud-init-helper"))
}

func cfZoneID(api *cloudflare.API, zoneID, zoneName string) (string, error) {
	if zoneID != "" {
		return zoneID, nil
	}

	return api.ZoneIDByName(zoneName)
}

func getRecords(ctx context.Context, api *cloudflare.API, zoneID, recordType, recordName, content string) ([]cloudflare.DNSRecord, error) {
	var res, records []cloudflare.DNSRecord
	var info *cloudflare.ResultInfo
	var err error

	zid := cloudflare.ZoneIdentifier(zoneID)
	p := cloudflare.ListDNSRecordsParams{
		Type:    recordType,
		Name:    recordName,
		Content: content,
	}

	for {
		if records, info, err = api.ListDNSRecords(ctx, zid, p); err != nil {
			return res, err
		}

		res = append(res, records...)

		if info.HasMorePages() {
			continue
		}

		return res, nil
	}
}

func createMXRecord(ctx context.Context, api *cloudflare.API, zoneID string, mailDomain string, mxHost string, weight int) error {
	zid := cloudflare.ZoneIdentifier(zoneID)
	priority := uint16(weight)
	params := cloudflare.CreateDNSRecordParams{
		Type:     "MX",
		Name:     mailDomain,
		Content:  mxHost,
		Priority: &priority,
	}
	_, err := api.CreateDNSRecord(ctx, zid, params)
	return err
}

func getRecord(ctx context.Context, api *cloudflare.API, zoneID string, rid string) (cloudflare.DNSRecord, error) {
	zid := cloudflare.ZoneIdentifier(zoneID)
	return api.GetDNSRecord(ctx, zid, rid)
}

func createRecord(ctx context.Context, api *cloudflare.API, zoneID string, rtype string, name string, content string) error {
	zid := cloudflare.ZoneIdentifier(zoneID)

	if rtype == "TXT" {
		content = ensureQuoted(content)
	}

	params := cloudflare.CreateDNSRecordParams{
		Type:    rtype,
		Name:    name,
		Content: content,
	}

	_, err := api.CreateDNSRecord(ctx, zid, params)
	return err
}

func updateRecord(ctx context.Context, api *cloudflare.API, zoneID string, rid string, rtype string, name string, content string) error {
	zid := cloudflare.ZoneIdentifier(zoneID)

	if rtype == "TXT" {
		content = ensureQuoted(content)
	}

	params := cloudflare.UpdateDNSRecordParams{
		ID:      rid,
		Type:    rtype,
		Name:    name,
		Content: content,
	}

	_, err := api.UpdateDNSRecord(ctx, zid, params)
	return err
}

func deleteRecord(ctx context.Context, api *cloudflare.API, zoneID string, rid string) error {
	zid := cloudflare.ZoneIdentifier(zoneID)
	return api.DeleteDNSRecord(ctx, zid, rid)
}

func ensureQuoted(s string) string {
	if len(s) > 0 {
		if s[0] != '"' {
			s = strconv.Quote(s)
		}
		if s[len(s)-1] != '"' {
			s = strconv.Quote(s)
		}
	}
	return s
}
