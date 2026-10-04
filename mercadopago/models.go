package mercadopago

import (
	"fmt"
	"time"
)

// PaymentSearchResponse represents the paginated response returned by MercadoPago's /v1/payments/search endpoint.
type PaymentSearchResponse struct {
	Paging  Paging    `json:"paging"`
	Results []Payment `json:"results"`
}

// Paging contains pagination metadata.
type Paging struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Payment represents a single transaction record from MercadoPago.
type Payment struct {
	ID                        int64                  `json:"id"`
	DateCreated               string                 `json:"date_created"`
	DateApproved              *string                `json:"date_approved"`
	DateLastUpdated           string                 `json:"date_last_updated"`
	DateOfExpiration          *string                `json:"date_of_expiration"`
	MoneyReleaseDate          *string                `json:"money_release_date"`
	MoneyReleaseStatus        string                 `json:"money_release_status"`
	OperationType             string                 `json:"operation_type"`
	IssuerID                  interface{}            `json:"issuer_id"`
	PaymentMethodID           string                 `json:"payment_method_id"`
	PaymentTypeID             string                 `json:"payment_type_id"`
	PaymentMethod             PaymentMethod          `json:"payment_method"`
	Status                    string                 `json:"status"`
	StatusDetail              string                 `json:"status_detail"`
	CurrencyID                string                 `json:"currency_id"`
	Description               string                 `json:"description"`
	LiveMode                  bool                   `json:"live_mode"`
	SponsorID                 interface{}            `json:"sponsor_id"`
	AuthorizationCode         interface{}            `json:"authorization_code"`
	MoneyReleaseSchema        interface{}            `json:"money_release_schema"`
	TaxesAmount               float64                `json:"taxes_amount"`
	CounterCurrency           interface{}            `json:"counter_currency"`
	BrandID                   interface{}            `json:"brand_id"`
	ShippingAmount            float64                `json:"shipping_amount"`
	BuildVersion              interface{}            `json:"build_version"`
	IdempotencyKey            interface{}            `json:"idempotency_key"`
	PosID                     interface{}            `json:"pos_id"`
	StoreID                   interface{}            `json:"store_id"`
	IntegratorID              interface{}            `json:"integrator_id"`
	PlatformID                interface{}            `json:"platform_id"`
	CorporationID             interface{}            `json:"corporation_id"`
	ChargesExecutionInfo      interface{}            `json:"charges_execution_info"`
	PayerID                   interface{}            `json:"payer_id"`
	Payer                     map[string]interface{} `json:"payer"`
	CollectorID               *int64                 `json:"collector_id"`
	Collector                 Collector              `json:"collector"`
	MarketplaceOwner          interface{}            `json:"marketplace_owner"`
	Metadata                  map[string]interface{} `json:"metadata"`
	AdditionalInfo            AdditionalInfo         `json:"additional_info"`
	Order                     interface{}            `json:"order"`
	ExternalReference         interface{}            `json:"external_reference"`
	TransactionAmount         float64                `json:"transaction_amount"`
	TransactionAmountRefunded float64                `json:"transaction_amount_refunded"`
	CouponAmount              float64                `json:"coupon_amount"`
	DifferentialPricingID     interface{}            `json:"differential_pricing_id"`
	FinancingGroup            interface{}            `json:"financing_group"`
	DeductionSchema           interface{}            `json:"deduction_schema"`
	Installments              int                    `json:"installments"`
	TransactionDetails        TransactionDetails     `json:"transaction_details"`
	FeeDetails                []FeeDetail            `json:"fee_details"`
	ChargesDetails            []ChargeDetail         `json:"charges_details"`
	Captured                  bool                   `json:"captured"`
	BinaryMode                bool                   `json:"binary_mode"`
	CallForAuthorizeID        interface{}            `json:"call_for_authorize_id"`
	StatementDescriptor       interface{}            `json:"statement_descriptor"`
	Card                      Card                   `json:"card"`
	NotificationURL           interface{}            `json:"notification_url"`
	Refunds                   []interface{}          `json:"refunds"`
	ProcessingMode            string                 `json:"processing_mode"`
	MerchantAccountID         interface{}            `json:"merchant_account_id"`
	MerchantNumber            interface{}            `json:"merchant_number"`
	AcquirerReconciliation    []interface{}          `json:"acquirer_reconciliation"`
	PointOfInteraction        *PointOfInteraction    `json:"point_of_interaction"`
	AccountsInfo              interface{}            `json:"accounts_info"`
	ReleaseInfo               interface{}            `json:"release_info"`
	Tags                      []string               `json:"tags"`
	TenantContext             interface{}            `json:"tenant_context"`
	ShippingCost              float64                `json:"shipping_cost"`
}

// PaymentMethod describes the payment method used.
type PaymentMethod struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	IssuerID interface{} `json:"issuer_id"`
}

// ChargesExecutionInfo contains internal execution timing and identifier.
type ChargesExecutionInfo struct {
	InternalExecution interface{} `json:"internal_execution"`
}

// InternalExecution holds execution date and execution ID.
type InternalExecution struct {
	Date        interface{} `json:"date"`
	ExecutionID interface{} `json:"execution_id"`
}

// Collector describes the merchant / collector receiving the funds.
type Collector struct {
	ID             int64          `json:"id"`
	OperatorID     interface{}    `json:"operator_id"`
	Email          interface{}    `json:"email"`
	Identification Identification `json:"identification"`
	Phone          interface{}    `json:"phone"`
	FirstName      interface{}    `json:"first_name"`
	LastName       interface{}    `json:"last_name"`
}

// Identification holds document number and type.
type Identification struct {
	Number interface{} `json:"number"`
	Type   interface{} `json:"type"`
}

// AdditionalInfo contains tracking and platform info.
type AdditionalInfo struct {
	TrackingID interface{} `json:"tracking_id"`
}

// Order represents related order info.
type Order struct {
	Type interface{} `json:"type"`
	ID   interface{} `json:"id"`
}

// TransactionDetails contains financial breakdown of the transaction.
type TransactionDetails struct {
	PaymentMethodReferenceID interface{} `json:"payment_method_reference_id"`
	AcquirerReference        interface{} `json:"acquirer_reference"`
	NetReceivedAmount        float64     `json:"net_received_amount"`
	TotalPaidAmount          float64     `json:"total_paid_amount"`
	OverpaidAmount           float64     `json:"overpaid_amount"`
	ExternalResourceURL      interface{} `json:"external_resource_url"`
	InstallmentAmount        float64     `json:"installment_amount"`
	FinancialInstitution     interface{} `json:"financial_institution"`
	PayableDeferralPeriod    interface{} `json:"payable_deferral_period"`
	BankTransferID           interface{} `json:"bank_transfer_id"`
	TransactionID            interface{} `json:"transaction_id"`
}

// FeeDetail represents fees charged on the payment.
type FeeDetail struct {
	Type     string      `json:"type"`
	Amount   float64     `json:"amount"`
	FeePayer interface{} `json:"fee_payer"`
}

// ChargeDetail specifies breakdown of taxes and fees applied.
type ChargeDetail struct {
	ID               interface{}            `json:"id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Accounts         interface{}            `json:"accounts"`
	ClientID         interface{}            `json:"client_id"`
	DateCreated      string                 `json:"date_created"`
	LastUpdated      string                 `json:"last_updated"`
	Amounts          interface{}            `json:"amounts"`
	Metadata         map[string]interface{} `json:"metadata"`
	ReserveID        interface{}            `json:"reserve_id"`
	ProcessingMode   string                 `json:"processing_mode"`
	ExternalChargeID interface{}            `json:"external_charge_id"`
	Rate             interface{}            `json:"rate"`
	BaseAmount       interface{}            `json:"base_amount"`
}

// ChargeAccounts describes source and destination accounts for the charge.
type ChargeAccounts struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ChargeAmounts details the original and refunded charge amounts.
type ChargeAmounts struct {
	Original float64 `json:"original"`
	Refunded float64 `json:"refunded"`
}

// Card contains payment card metadata.
type Card struct {
	ID              interface{} `json:"id"`
	FirstSixDigits  interface{} `json:"first_six_digits"`
	LastFourDigits  interface{} `json:"last_four_digits"`
	Bin             interface{} `json:"bin"`
	ExpirationMonth interface{} `json:"expiration_month"`
	ExpirationYear  interface{} `json:"expiration_year"`
	DateCreated     interface{} `json:"date_created"`
	DateLastUpdated interface{} `json:"date_last_updated"`
	Country         interface{} `json:"country"`
	Tags            interface{} `json:"tags"`
	Cardholder      interface{} `json:"cardholder"`
}

// Cardholder contains the cardholder name and identification.
type Cardholder struct {
	Name           interface{} `json:"name"`
	Identification interface{} `json:"identification"`
}

// ReleaseInfo holds details about advance provider and release events.
type ReleaseInfo struct {
	AdvanceProvider         interface{} `json:"advance_provider"`
	AdvanceProviderUser     interface{} `json:"advance_provider_user"`
	Events                  interface{} `json:"events"`
	NewPaymentsReleaseModel interface{} `json:"new_payments_release_model"`
}

// ReleaseEvent specifies a release or refund event.
type ReleaseEvent struct {
	EntityID interface{} `json:"entity_id"`
	ID       interface{} `json:"id"`
	Type     interface{} `json:"type"`
}

// PointOfInteraction holds POS, in-store, or online interaction details.
type PointOfInteraction struct {
	Type            interface{}                 `json:"type"`
	BusinessInfo    BusinessInfo                `json:"business_info"`
	Location        interface{}                 `json:"location"`
	References      interface{}                 `json:"references"`
	TransactionData *InteractionTransactionData `json:"transaction_data"`
}

// BusinessInfo details the sub-unit and branch.
type BusinessInfo struct {
	Unit    string `json:"unit"`
	SubUnit string `json:"sub_unit"`
	Branch  string `json:"branch"`
}

// Location describes store geolocation and state.
type Location struct {
	StateID   interface{} `json:"state_id"`
	Source    interface{} `json:"source"`
	Latitude  interface{} `json:"latitude"`
	Longitude interface{} `json:"longitude"`
}

// InteractionReference holds reference entity IDs.
type InteractionReference struct {
	ID   interface{} `json:"id"`
	Type interface{} `json:"type"`
}

// InteractionTransactionData contains QR, bank, or subscription details.
type InteractionTransactionData struct {
	QRCode               interface{} `json:"qr_code"`
	BankTransferID       interface{} `json:"bank_transfer_id"`
	TransactionID        interface{} `json:"transaction_id"`
	E2EID                interface{} `json:"e2e_id"`
	FinancialInstitution interface{} `json:"financial_institution"`
	TicketURL            interface{} `json:"ticket_url"`
	MerchantCategoryCode interface{} `json:"merchant_category_code"`
	IsEndConsumer        interface{} `json:"is_end_consumer"`
}

// IsIncome returns true if the payment is an income for the user, false if it is an expense.
func (p *Payment) IsIncome(myUserID int64) bool {
	isCollector := (p.Collector.ID > 0 && p.Collector.ID == myUserID) ||
		(p.CollectorID != nil && *p.CollectorID == myUserID)

	return (myUserID > 0 && isCollector) ||
		(p.PointOfInteraction != nil && p.PointOfInteraction.BusinessInfo.SubUnit == "money_inflows") ||
		p.OperationType == "account_fund"
}

// GetTransactionType determines whether the payment is an Income, Expense, or Refund relative to the account owner.
func (p *Payment) GetTransactionType(myUserID int64) string {
	if p.Status == "refunded" || p.NetAmount() == 0 {
		return "🔄 Reintegro"
	}

	if p.IsIncome(myUserID) {
		return "🟢 Ingreso"
	}

	return "🔴 Egreso"
}

// IsExecuted returns "Si" only if the payment was approved or refunded, otherwise returns empty string.
func (p *Payment) IsExecuted() string {
	if p.Status == "approved" || p.Status == "refunded" {
		return "Si"
	}
	return ""
}

// GetOriginDestinationMethod returns the relevant funding source ID.
func (p *Payment) GetOriginDestinationMethod() string {
	if p.PaymentTypeID != "" {
		return p.PaymentTypeID
	}
	if p.PaymentMethodID != "" {
		return p.PaymentMethodID
	}
	return "unknown"
}

// GetPaymentMethodSpanish returns the payment method translated to Spanish.
func (p *Payment) GetPaymentMethodSpanish() string {
	switch p.PaymentTypeID {
	case "account_money":
		return "Dinero en cuenta"
	case "bank_transfer":
		return "Transferencia bancaria"
	case "credit_card":
		return "Tarjeta de crédito"
	case "debit_card":
		return "Tarjeta de débito"
	case "prepaid_card":
		return "Tarjeta prepaga"
	case "ticket":
		return "Efectivo"
	case "atm":
		return "Cajero automático"
	default:
		switch p.PaymentMethodID {
		case "account_money":
			return "Dinero en cuenta"
		case "cvu":
			return "Transferencia bancaria"
		default:
			if p.PaymentTypeID != "" {
				return p.PaymentTypeID
			}
			return "Otro"
		}
	}
}

// parseTime parses the payment's ISO8601 creation date.
func (p *Payment) parseTime() time.Time {
	if p.DateCreated == "" {
		return time.Time{}
	}
	layouts := []string{
		"2006-01-02T15:04:05.000-07:00",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, p.DateCreated); err == nil {
			return t
		}
	}
	return time.Time{}
}

// FormattedDate returns the creation date in DD/MM/YYYY format matching Argentine locale and spreadsheet template.
func (p *Payment) FormattedDate() string {
	t := p.parseTime()
	if t.IsZero() {
		return ""
	}
	return t.Format("02/01/2006")
}

// FormattedTime returns the creation time in HH:MM:SS format.
func (p *Payment) FormattedTime() string {
	t := p.parseTime()
	if t.IsZero() {
		return ""
	}
	return t.Format("15:04:05")
}

// NetAmount calculates the actual net amount after deducting refunds.
func (p *Payment) NetAmount() float64 {
	net := p.TransactionAmount - p.TransactionAmountRefunded
	if net < 0 {
		return 0
	}
	return net
}

// RefundNotes returns the refund description if refunded, or empty string if regular payment.
func (p *Payment) RefundNotes() string {
	if p.TransactionAmountRefunded > 0 {
		if p.NetAmount() == 0 {
			return "Reembolso total"
		}
		return fmt.Sprintf("Reembolso parcial de %.2f", p.TransactionAmountRefunded)
	}
	return ""
}

// FormattedNotes returns the description, annotating refund status if applicable.
func (p *Payment) FormattedNotes() string {
	if p.TransactionAmountRefunded > 0 {
		if p.NetAmount() == 0 {
			return fmt.Sprintf("%s (Reembolso total)", p.Description)
		}
		return fmt.Sprintf("%s (Reembolso parcial de %.2f)", p.Description, p.TransactionAmountRefunded)
	}
	return p.Description
}



