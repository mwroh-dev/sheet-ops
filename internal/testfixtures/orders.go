package testfixtures

import "github.com/xuri/excelize/v2"

func CreateOrdersFixture(path string) error {
	file := excelize.NewFile()
	const sheet = "Orders"
	file.SetSheetName("Sheet1", sheet)
	rows := [][]any{
		{"order_id", "customer", "amount"},
		{1, "A", 50000},
		{2, "B", 125000},
		{3, "C", 140000},
		{4, "D", 90000},
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err := file.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}
	return file.SaveAs(path)
}

func CreateSummaryOrdersFixture(path string) error {
	file := excelize.NewFile()
	const sheet = "주문내역"
	file.SetSheetName("Sheet1", sheet)
	rows := [][]any{
		{"주문번호", "상품명", "배송상태", "결제금액"},
		{1001, "사과", "배송완료", 10000},
		{1002, "배", "취소", 20000},
		{1003, "사과", "배송완료", 30000},
		{1004, "배", "배송완료", 40000},
		{1005, "사과", "배송중", 5000},
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err := file.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}
	return file.SaveAs(path)
}
