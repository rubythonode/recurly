package recurly

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
)

var _ InvoicesService = &invoicesImpl{}

// invoicesImpl handles communication with theinvoices related methods
// of the recurly API.
type invoicesImpl struct {
	client *Client
}

// NewInvoicesImpl returns a new instance of invoicesImpl.
func NewInvoicesImpl(client *Client) *invoicesImpl {
	return &invoicesImpl{client: client}
}

// List returns a list of all invoices.
// https://dev.recurly.com/docs/list-invoices
func (s *invoicesImpl) List(params Params) (*Response, []Invoice, error) {
	req, err := s.client.newRequest("GET", "invoices", params, nil)
	if err != nil {
		return nil, nil, err
	}

	var p struct {
		XMLName  xml.Name  `xml:"invoices"`
		Invoices []Invoice `xml:"invoice"`
	}
	resp, err := s.client.do(req, &p)

	return resp, p.Invoices, err
}

// ListAccount returns a list of all invoices for an account.
// https://dev.recurly.com/docs/list-an-accounts-invoices
func (s *invoicesImpl) ListAccount(accountCode string, params Params) (*Response, []Invoice, error) {
	action := fmt.Sprintf("accounts/%s/invoices", accountCode)
	req, err := s.client.newRequest("GET", action, params, nil)
	if err != nil {
		return nil, nil, err
	}

	var p struct {
		XMLName  xml.Name  `xml:"invoices"`
		Invoices []Invoice `xml:"invoice"`
	}
	resp, err := s.client.do(req, &p)

	return resp, p.Invoices, err
}

// Get returns detailed information about an invoice including line items and
// payments. Transactions returned with the invoice are sorted from oldest to
// newest.
// https://dev.recurly.com/docs/lookup-invoice-details
func (s *invoicesImpl) Get(invoiceNumber int) (*Response, *Invoice, error) {
	action := fmt.Sprintf("invoices/%d", invoiceNumber)
	req, err := s.client.newRequest("GET", action, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		return resp, nil, err
	}

	// Sort transactions.
	sort.Sort(Transactions(dst.Transactions))

	return resp, &dst, err
}

// GetPDF retrieves the invoice as a PDF.
// The language parameters allows you to specify a language to translate the
// invoice into. If empty, English will be used. Options: Danish, German,
// Spanish, French, Hindi, Japanese, Dutch, Portuguese, Russian, Turkish, Chinese.
// https://dev.recurly.com/docs/retrieve-a-pdf-invoice
func (s *invoicesImpl) GetPDF(invoiceNumber int, language string) (*Response, *bytes.Buffer, error) {
	action := fmt.Sprintf("invoices/%d", invoiceNumber)
	req, err := s.client.newRequest("GET", action, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	if language == "" {
		language = "English"
	}

	req.Header.Set("Accept", "application/pdf")
	req.Header.Set("Accept-Language", language)

	var pdf bytes.Buffer
	resp, err := s.client.do(req, &pdf)

	return resp, &pdf, err
}

// Preview allows you to display the invoice details, including estimated tax,
// before you post it.
// https://dev.recurly.com/docs/post-an-invoice-invoice-pending-charges-on-an-acco
func (s *invoicesImpl) Preview(accountCode string) (*Response, *Invoice, error) {
	action := fmt.Sprintf("accounts/%s/invoices/preview", accountCode)
	req, err := s.client.newRequest("POST", action, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)

	return resp, &dst, err
}

// Create posts an accounts pending charges to a new invoice on that account.
// When you post one-time charges to an account, these will remain pending
// until they are invoiced. An account is automatically invoiced when the
// subscription renews. However, there are times when it is appropriate to
// invoice an account before the renewal. If the subscriber has a yearly
// subscription, you might want to collect the one-time charges well before the renewal.
// https://dev.recurly.com/docs/post-an-invoice-invoice-pending-charges-on-an-acco
func (s *invoicesImpl) Create(accountCode string, invoice Invoice) (*Response, *Invoice, error) {
	action := fmt.Sprintf("accounts/%s/invoices", accountCode)
	req, err := s.client.newRequest("POST", action, nil, invoice)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)

	return resp, &dst, err
}

// Collect force retries the card on file for the invoice.
// Allows to collect a past-due or pending invoice. This API is rate limited
// and only one collection attempt per account is allowed within an hour.
// Will return status code 400 if rate limit hit or if invoice not in the
// correct state.
// https://dev.recurly.com/v2.5/docs/collect-an-invoice
func (s *invoicesImpl) Collect(invoiceNumber int) (*Response, *Invoice, error) {
	action := fmt.Sprintf("invoices/%d/collect", invoiceNumber)
	req, err := s.client.newRequest("PUT", action, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		return resp, nil, err
	}

	return resp, &dst, err
}

// MarkPaid marks an invoice as paid successfully.
// https://dev.recurly.com/docs/mark-an-invoice-as-paid-successfully
func (s *invoicesImpl) MarkPaid(invoiceNumber int) (*Response, *Invoice, error) {
	action := fmt.Sprintf("invoices/%d/mark_successful", invoiceNumber)
	req, err := s.client.newRequest("PUT", action, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)

	return resp, &dst, err
}

// MarkFailed marks an invoice as failed.
// https://dev.recurly.com/docs/mark-an-invoice-as-failed-collection
func (s *invoicesImpl) MarkFailed(invoiceNumber int) (*Response, *Invoice, error) {
	action := fmt.Sprintf("invoices/%d/mark_failed", invoiceNumber)
	req, err := s.client.newRequest("PUT", action, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)

	return resp, &dst, err
}

// RefundVoidOpenAmount allows custom invoice amounts to be refunded and generates a refund invoice.
// Full open amount refunds of invoices with an unsettled transaction will void
// the transaction and generate a void invoice.
// https://dev.recurly.com/docs/line-item-refunds
func (s *invoicesImpl) RefundVoidOpenAmount(invoiceNumber int, amountInCents int, refundApplyOrder string) (*Response, *Invoice, error) {
	action := fmt.Sprintf("invoices/%d/refund", invoiceNumber)
	data := struct {
		XMLName          xml.Name `xml:"invoice"`
		AmountInCents    int      `xml:"amount_in_cents,omitempty"`
		RefundApplyOrder string   `xml:"refund_apply_order,omitempty"`
	}{
		AmountInCents:    amountInCents,    // Amount is required
		RefundApplyOrder: refundApplyOrder, // Refund Apply Order defaults to "credit" (vs "transaction")
	}
	req, err := s.client.newRequest("POST", action, nil, data)
	if err != nil {
		return nil, nil, err
	}

	var dst Invoice
	resp, err := s.client.do(req, &dst)

	return resp, &dst, err
}
