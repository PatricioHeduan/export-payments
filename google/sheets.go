package google

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"export-payments/config"
	"export-payments/mercadopago"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient defines the contract for interacting with Google Sheets API across separate sheets.
type SheetsClient interface {
	AppendPayments(ctx context.Context, payments []mercadopago.Payment) error
	EnsureHeaderRows(ctx context.Context) error
}

type sheetsClient struct {
	service          *sheets.Service
	spreadsheetID    string
	incomeSheetName  string
	expenseSheetName string
	userID           int64
}

// NewSheetsClient instantiates a Google Sheets client configured for separate Income and Expense tabs.
func NewSheetsClient(ctx context.Context, cfg *config.Config) (SheetsClient, error) {
	var opts []option.ClientOption

	if cfg.GoogleCredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.GoogleCredentialsJSON)))
	} else if cfg.GoogleCredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.GoogleCredentialsPath))
	}

	opts = append(opts, option.WithScopes(sheets.SpreadsheetsScope))

	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Sheets service: %w", err)
	}

	return &sheetsClient{
		service:          srv,
		spreadsheetID:    cfg.SpreadsheetID,
		incomeSheetName:  cfg.IncomeSheetName,
		expenseSheetName: cfg.ExpenseSheetName,
		userID:           cfg.MPUserID,
	}, nil
}

// IncomeHeaders defines the columns for the Income sheet matching the template table.
var IncomeHeaders = []interface{}{
	"Fecha",
	"Hora",
	"Categoría",
	"Descripción",
	"Monto ($)",
	"Medio de cobro",
	"Notas",
	"ID Transacción",
}

// ExpenseHeaders defines the columns for the Expense sheet matching the template table.
var ExpenseHeaders = []interface{}{
	"Fecha",
	"Hora",
	"Categoría",
	"Descripción",
	"Monto ($)",
	"Medio de pago",
	"Tipo",
	"Notas",
	"ID Transacción",
}

// colIndexToLetter converts a 0-indexed column number to standard spreadsheet A1 column letter (0->A, 8->I, 26->AA).
func colIndexToLetter(colIdx int) string {
	letter := ""
	colIdx++
	for colIdx > 0 {
		rem := (colIdx - 1) % 26
		letter = string(rune('A'+rem)) + letter
		colIdx = (colIdx - 1) / 26
	}
	return letter
}

// columnMapping stores the dynamic 0-based column indices discovered from the sheet's header row.
type columnMapping struct {
	fechaCol     int
	horaCol      int
	catCol       int
	descCol      int
	montoCol     int
	medioPagoCol int
	tipoCol      int
	notasCol     int
	idCol        int
	maxCol       int
}

// detectColumns inspects the sheet's header row and dynamically maps field names to column indices.
func detectColumns(headerRow []interface{}, isIncome bool) columnMapping {
	// Sensible defaults matching the template table
	cm := columnMapping{
		fechaCol:     0,
		horaCol:      1,
		catCol:       2,
		descCol:      3,
		montoCol:     4,
		medioPagoCol: 5,
		tipoCol:      -1,
		notasCol:     6,
		idCol:        7,
	}
	if !isIncome {
		cm.tipoCol = 6
		cm.notasCol = 7
		cm.idCol = 8
	}

	for idx, cell := range headerRow {
		c := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", cell)))
		if strings.Contains(c, "id") || strings.Contains(c, "transacci") {
			cm.idCol = idx
		} else if strings.Contains(c, "monto") || strings.Contains(c, "amount") || strings.Contains(c, "importe") {
			cm.montoCol = idx
		} else if strings.Contains(c, "descrip") || strings.Contains(c, "detalle") {
			cm.descCol = idx
		} else if strings.Contains(c, "categor") {
			cm.catCol = idx
		} else if strings.Contains(c, "tipo") || strings.Contains(c, "type") {
			cm.tipoCol = idx
		} else if strings.Contains(c, "nota") || strings.Contains(c, "notes") {
			cm.notasCol = idx
		} else if strings.Contains(c, "medio") || strings.Contains(c, "pago") || strings.Contains(c, "cobro") {
			cm.medioPagoCol = idx
		} else if strings.Contains(c, "fecha") || strings.Contains(c, "date") {
			cm.fechaCol = idx
		} else if strings.Contains(c, "hora") || strings.Contains(c, "time") {
			cm.horaCol = idx
		}
	}

	cols := []int{cm.fechaCol, cm.horaCol, cm.catCol, cm.descCol, cm.montoCol, cm.medioPagoCol, cm.tipoCol, cm.notasCol, cm.idCol}
	max := len(headerRow) - 1
	for _, col := range cols {
		if col > max {
			max = col
		}
	}
	cm.maxCol = max
	return cm
}

// EnsureHeaderRows verifies if headers exist on both sheets; if empty, sets default headers.
// If headers exist but lacks "ID Transacción", adds it to the adjacent column.
func (s *sheetsClient) EnsureHeaderRows(ctx context.Context) error {
	if err := s.ensureHeaderRowForSheet(ctx, s.incomeSheetName, IncomeHeaders); err != nil {
		return fmt.Errorf("failed ensuring headers for income sheet (%s): %w", s.incomeSheetName, err)
	}
	if err := s.ensureHeaderRowForSheet(ctx, s.expenseSheetName, ExpenseHeaders); err != nil {
		return fmt.Errorf("failed ensuring headers for expense sheet (%s): %w", s.expenseSheetName, err)
	}
	return nil
}

func (s *sheetsClient) ensureHeaderRowForSheet(ctx context.Context, sheetName string, defaultHeaders []interface{}) error {
	readResp, err := s.service.Spreadsheets.Values.Get(s.spreadsheetID, sheetName+"!A1:Z1").Context(ctx).Do()
	if err != nil {
		return err
	}

	if len(readResp.Values) == 0 || len(readResp.Values[0]) == 0 {
		log.Printf("[GoogleSheets] Sheet '%s' is empty. Writing default header row...", sheetName)
		headerRange := &sheets.ValueRange{
			Values: [][]interface{}{defaultHeaders},
		}
		targetRange := fmt.Sprintf("%s!A1:%s1", sheetName, colIndexToLetter(len(defaultHeaders)-1))
		_, err := s.service.Spreadsheets.Values.Update(s.spreadsheetID, targetRange, headerRange).
			ValueInputOption("USER_ENTERED").
			Context(ctx).
			Do()
		if err != nil {
			return err
		}
		log.Printf("[GoogleSheets] Default header row written successfully to '%s'.", sheetName)
		return nil
	}

	// Verify if "ID Transacción" exists in row 1; if missing, add it to avoid column collision
	hasIDCol := false
	for _, cell := range readResp.Values[0] {
		cellStr := strings.ToLower(fmt.Sprintf("%v", cell))
		if strings.Contains(cellStr, "id") || strings.Contains(cellStr, "transacci") {
			hasIDCol = true
			break
		}
	}
	if !hasIDCol {
		nextColIdx := len(readResp.Values[0])
		colLetter := colIndexToLetter(nextColIdx)
		cellRange := fmt.Sprintf("%s!%s1", sheetName, colLetter)
		valRange := &sheets.ValueRange{
			Values: [][]interface{}{{"ID Transacción"}},
		}
		_, err := s.service.Spreadsheets.Values.Update(s.spreadsheetID, cellRange, valRange).
			ValueInputOption("USER_ENTERED").
			Context(ctx).
			Do()
		if err != nil {
			log.Printf("[GoogleSheets] [%s] Warning: could not add ID Transacción header: %v", sheetName, err)
		} else {
			log.Printf("[GoogleSheets] [%s] Added 'ID Transacción' header at %s1", sheetName, colLetter)
		}
	}

	return nil
}

type existingRowInfo struct {
	sheetRowNum int
	amount      float64
	category    string
	tipo        string
	notes       string
}

// AppendPayments splits incoming transactions into Income and Expense categories and routes them to their respective sheets.
func (s *sheetsClient) AppendPayments(ctx context.Context, payments []mercadopago.Payment) error {
	if len(payments) == 0 {
		log.Println("[GoogleSheets] No transactions to process.")
		return nil
	}

	var incomePayments []mercadopago.Payment
	var expensePayments []mercadopago.Payment

	for _, p := range payments {
		if p.IsIncome(s.userID) {
			incomePayments = append(incomePayments, p)
		} else {
			expensePayments = append(expensePayments, p)
		}
	}

	log.Printf("[GoogleSheets] Routing %d income(s) to '%s' and %d expense(s) to '%s'...",
		len(incomePayments), s.incomeSheetName, len(expensePayments), s.expenseSheetName)

	if err := s.processSheet(ctx, s.incomeSheetName, incomePayments, true); err != nil {
		return fmt.Errorf("failed processing income sheet: %w", err)
	}

	if err := s.processSheet(ctx, s.expenseSheetName, expensePayments, false); err != nil {
		return fmt.Errorf("failed processing expense sheet: %w", err)
	}

	return nil
}

// processSheet reads existing rows, deduplicates, updates refunds, and writes new rows directly inside the table.
func (s *sheetsClient) processSheet(ctx context.Context, sheetName string, payments []mercadopago.Payment, isIncome bool) error {
	if len(payments) == 0 {
		log.Printf("[GoogleSheets] No payments to process for sheet '%s'.", sheetName)
		return nil
	}

	// 1. Fetch current sheet content to inspect headers, existing IDs, amounts, and filled table rows
	readResp, err := s.service.Spreadsheets.Values.Get(s.spreadsheetID, sheetName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet '%s': %w", sheetName, err)
	}

	var headerRow []interface{}
	if len(readResp.Values) > 0 {
		headerRow = readResp.Values[0]
	}

	cm := detectColumns(headerRow, isIncome)
	log.Printf("[GoogleSheets] [%s] Column mapping: Fecha=%d, Hora=%d, Categoria=%d, Descripcion=%d, Monto=%d, MedioPago=%d, Tipo=%d, Notas=%d, ID=%d (maxCol=%d)",
		sheetName, cm.fechaCol, cm.horaCol, cm.catCol, cm.descCol, cm.montoCol, cm.medioPagoCol, cm.tipoCol, cm.notasCol, cm.idCol, cm.maxCol)

	existingRows := make(map[string]existingRowInfo)
	lastFilledRow := 1
	consecutiveEmpty := 0

	if len(readResp.Values) > 1 {
		for i, row := range readResp.Values[1:] {
			sheetRowNum := i + 2

			// Check if row has any non-empty cell
			rowHasContent := false
			for _, cell := range row {
				if cell != nil && strings.TrimSpace(fmt.Sprintf("%v", cell)) != "" {
					rowHasContent = true
					break
				}
			}

			if rowHasContent {
				// If we encounter 10+ consecutive empty rows earlier, stop advancing lastFilledRow
				// to avoid jumping to rogue entries that were appended at the bottom of the sheet.
				if consecutiveEmpty < 10 {
					lastFilledRow = sheetRowNum
					consecutiveEmpty = 0

					// Extract ID if available
					var idVal string
					if len(row) > cm.idCol {
						idVal = strings.TrimSpace(fmt.Sprintf("%v", row[cm.idCol]))
					}
					// Fallback search in any cell if ID was placed in another column previously
					if idVal == "" {
						for _, c := range row {
							str := strings.TrimSpace(fmt.Sprintf("%v", c))
							if len(str) >= 10 && isNumeric(str) {
								idVal = str
								break
							}
						}
					}

					if idVal != "" {
						var currentAmount float64
						if len(row) > cm.montoCol {
							currentAmount = parseSheetAmount(row[cm.montoCol])
						}
						var category, tipo, notes string
						if cm.catCol >= 0 && len(row) > cm.catCol {
							category = fmt.Sprintf("%v", row[cm.catCol])
						}
						if cm.tipoCol >= 0 && len(row) > cm.tipoCol {
							tipo = fmt.Sprintf("%v", row[cm.tipoCol])
						}
						if cm.notasCol >= 0 && len(row) > cm.notasCol {
							notes = fmt.Sprintf("%v", row[cm.notasCol])
						}

						existingRows[idVal] = existingRowInfo{
							sheetRowNum: sheetRowNum,
							amount:      currentAmount,
							category:    category,
							tipo:        tipo,
							notes:       notes,
						}
					}
				}
			} else {
				consecutiveEmpty++
			}
		}
	}

	log.Printf("[GoogleSheets] [%s] Last active table row: %d (found %d existing IDs)",
		sheetName, lastFilledRow, len(existingRows))

	// 2. Process payments: update changed refunds, or collect new unique rows
	var newRows [][]interface{}
	for _, p := range payments {
		idStr := strconv.FormatInt(p.ID, 10)
		netAmount := p.NetAmount()
		refundNote := p.RefundNotes()

		// Case A: Transaction already exists in this sheet
		if existing, exists := existingRows[idStr]; exists {
			if p.TransactionAmountRefunded > 0 && existing.amount != netAmount {
				log.Printf("[GoogleSheets] [%s] Updating transaction %s at row %d: refund detected (old: %.2f, new net: %.2f)",
					sheetName, idStr, existing.sheetRowNum, existing.amount, netAmount)

				updateRow := make([]interface{}, cm.maxCol+1)
				updateRow[cm.fechaCol] = p.FormattedDate()
				if cm.horaCol >= 0 {
					updateRow[cm.horaCol] = p.FormattedTime()
				}
				if cm.catCol >= 0 {
					updateRow[cm.catCol] = existing.category // Preserve user manual category
				}
				if cm.descCol >= 0 {
					updateRow[cm.descCol] = p.Description
				}
				if cm.montoCol >= 0 {
					updateRow[cm.montoCol] = netAmount
				}
				if cm.medioPagoCol >= 0 {
					updateRow[cm.medioPagoCol] = p.GetPaymentMethodSpanish()
				}
				if cm.tipoCol >= 0 {
					updateRow[cm.tipoCol] = existing.tipo // Preserve user manual type
				}
				if cm.notasCol >= 0 {
					updateRow[cm.notasCol] = refundNote
				}
				if cm.idCol >= 0 {
					updateRow[cm.idCol] = idStr
				}

				endColLetter := colIndexToLetter(cm.maxCol)
				updateRange := fmt.Sprintf("%s!A%d:%s%d", sheetName, existing.sheetRowNum, endColLetter, existing.sheetRowNum)
				updateVR := &sheets.ValueRange{
					Values: [][]interface{}{updateRow},
				}
				_, updateErr := s.service.Spreadsheets.Values.Update(s.spreadsheetID, updateRange, updateVR).
					ValueInputOption("USER_ENTERED").
					Context(ctx).
					Do()
				if updateErr != nil {
					log.Printf("[GoogleSheets] [%s] Warning: failed to update row %d for payment %s: %v",
						sheetName, existing.sheetRowNum, idStr, updateErr)
				}
			} else {
			}
			continue
		}

		// Case B: Transaction is brand new
		// Skip fully refunded transactions before initial export ($0 net)
		if netAmount == 0 && p.TransactionAmountRefunded >= p.TransactionAmount {
			continue
		}

		// Skip rejected or non-executed transactions
		if p.Status != "approved" && p.Status != "refunded" {
			continue
		}

		newRow := make([]interface{}, cm.maxCol+1)
		newRow[cm.fechaCol] = p.FormattedDate()
		if cm.horaCol >= 0 {
			newRow[cm.horaCol] = p.FormattedTime()
		}
		if cm.catCol >= 0 {
			newRow[cm.catCol] = "" // Blank for manual user categorization
		}
		if cm.descCol >= 0 {
			newRow[cm.descCol] = p.Description
		}
		if cm.montoCol >= 0 {
			newRow[cm.montoCol] = netAmount
		}
		if cm.medioPagoCol >= 0 {
			newRow[cm.medioPagoCol] = p.GetPaymentMethodSpanish()
		}
		if cm.tipoCol >= 0 {
			newRow[cm.tipoCol] = "" // Blank for manual user type
		}
		if cm.notasCol >= 0 {
			newRow[cm.notasCol] = refundNote
		}
		if cm.idCol >= 0 {
			newRow[cm.idCol] = idStr
		}

		newRows = append(newRows, newRow)
		existingRows[idStr] = existingRowInfo{amount: netAmount} // Avoid batch duplicates
	}

	// 3. Insert new unique rows directly into the table starting immediately after lastFilledRow
	if len(newRows) == 0 {
		log.Printf("[GoogleSheets] [%s] No new transactions to insert.", sheetName)
		return nil
	}

	startRow := lastFilledRow + 1
	endRow := startRow + len(newRows) - 1
	endColLetter := colIndexToLetter(cm.maxCol)
	writeRange := fmt.Sprintf("%s!A%d:%s%d", sheetName, startRow, endColLetter, endRow)

	valueRange := &sheets.ValueRange{
		Values: newRows,
	}

	// Using Values.Update writes directly into the formatted table rows, preserving table styles and formulas.
	resp, err := s.service.Spreadsheets.Values.Update(s.spreadsheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to write rows to '%s' at %s: %w", sheetName, writeRange, err)
	}

	log.Printf("[GoogleSheets] [%s] Successfully inserted %d new rows into table (updated range: %s)",
		sheetName, len(newRows), resp.UpdatedRange)
	return nil
}

// isNumeric checks if a string consists purely of digits.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// parseSheetAmount parses amounts from spreadsheet cells handling numbers, currency symbols, and commas.
func parseSheetAmount(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		clean := strings.ReplaceAll(strings.TrimSpace(v), "$", "")
		clean = strings.ReplaceAll(clean, " ", "")
		if strings.Contains(clean, ",") && strings.Contains(clean, ".") {
			clean = strings.ReplaceAll(clean, ".", "")
			clean = strings.ReplaceAll(clean, ",", ".")
		} else if strings.Contains(clean, ",") {
			clean = strings.ReplaceAll(clean, ",", ".")
		}
		f, _ := strconv.ParseFloat(clean, 64)
		return f
	default:
		return 0
	}
}
