package money

import "strings"

type Category struct {
	code   string
	name   string
	txType TransactionType
}

func NewCategory(txType TransactionType, code, name string) (Category, error) {
	code = strings.TrimSpace(strings.ToLower(code))
	name = strings.TrimSpace(name)
	if code == "" {
		code = "other"
	}
	if name == "" {
		name = code
	}
	category := Category{code: code, name: name, txType: txType}
	if err := category.ValidateFor(txType); err != nil {
		return Category{}, err
	}
	return category, nil
}

func Categorize(txType TransactionType, preferred, note string) (Category, error) {
	if category, ok := findCategory(txType, preferred); ok {
		return category, nil
	}
	if category, ok := inferCategoryFromNote(txType, note); ok {
		return category, nil
	}
	return defaultCategory(txType)
}

func (c Category) ValidateFor(txType TransactionType) error {
	if !txType.IsValid() {
		return ErrInvalidTransaction("transaction type must be income or expense")
	}
	if c.txType != txType {
		return ErrInvalidTransaction("category type does not match transaction type")
	}
	if c.code == "" || c.name == "" {
		return ErrInvalidTransaction("category is required")
	}
	return nil
}

func (c Category) Code() string {
	return c.code
}

func (c Category) Name() string {
	return c.name
}

func (c Category) Type() TransactionType {
	return c.txType
}

func (c Category) String() string {
	return c.name
}

func AvailableCategories(txType TransactionType) []Category {
	categories := categoryCatalog(txType)
	copied := make([]Category, len(categories))
	copy(copied, categories)
	return copied
}

func FindCategory(txType TransactionType, value string) (Category, bool) {
	return findCategory(txType, value)
}

func defaultCategory(txType TransactionType) (Category, error) {
	return NewCategory(txType, "other", "อื่น ๆ")
}

func findCategory(txType TransactionType, value string) (Category, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "#")
	if value == "" {
		return Category{}, false
	}
	for _, category := range categoryCatalog(txType) {
		if category.code == value || strings.ToLower(category.name) == value {
			return category, true
		}
	}
	return Category{}, false
}

func inferCategoryFromNote(txType TransactionType, note string) (Category, bool) {
	note = strings.TrimSpace(strings.ToLower(note))
	if note == "" {
		return Category{}, false
	}
	for _, rule := range categoryRules(txType) {
		for _, keyword := range rule.keywords {
			if strings.Contains(note, strings.ToLower(keyword)) {
				if category, ok := findCategory(txType, rule.code); ok {
					return category, true
				}
			}
		}
	}
	return Category{}, false
}

func categoryCatalog(txType TransactionType) []Category {
	switch txType {
	case TransactionTypeIncome:
		return []Category{
			mustCategory(txType, "salary", "เงินเดือน"),
			mustCategory(txType, "freelance", "ฟรีแลนซ์"),
			mustCategory(txType, "business", "ธุรกิจ"),
			mustCategory(txType, "investment", "การลงทุน"),
			mustCategory(txType, "bonus", "โบนัส"),
			mustCategory(txType, "gift", "ของขวัญ"),
			mustCategory(txType, "refund", "เงินคืน"),
			mustCategory(txType, "other", "อื่น ๆ"),
		}
	case TransactionTypeExpense:
		return []Category{
			mustCategory(txType, "food", "อาหาร"),
			mustCategory(txType, "drink", "เครื่องดื่ม"),
			mustCategory(txType, "transport", "เดินทาง"),
			mustCategory(txType, "shopping", "ช้อปปิ้ง"),
			mustCategory(txType, "housing", "ที่อยู่อาศัย"),
			mustCategory(txType, "utilities", "ค่าสาธารณูปโภค"),
			mustCategory(txType, "health", "สุขภาพ"),
			mustCategory(txType, "education", "การศึกษา"),
			mustCategory(txType, "entertainment", "บันเทิง"),
			mustCategory(txType, "travel", "ท่องเที่ยว"),
			mustCategory(txType, "family", "ครอบครัว"),
			mustCategory(txType, "debt", "หนี้สิน"),
			mustCategory(txType, "other", "อื่น ๆ"),
		}
	default:
		return nil
	}
}

type categoryRule struct {
	code     string
	keywords []string
}

func categoryRules(txType TransactionType) []categoryRule {
	switch txType {
	case TransactionTypeIncome:
		return []categoryRule{
			{code: "salary", keywords: []string{"เงินเดือน", "salary", "payroll"}},
			{code: "freelance", keywords: []string{"ฟรีแลนซ์", "freelance", "รับงาน", "จ้าง"}},
			{code: "business", keywords: []string{"ขาย", "ธุรกิจ", "business", "ลูกค้า"}},
			{code: "investment", keywords: []string{"ปันผล", "หุ้น", "กองทุน", "ดอกเบี้ย", "investment"}},
			{code: "bonus", keywords: []string{"โบนัส", "bonus"}},
			{code: "gift", keywords: []string{"ของขวัญ", "gift", "ให้"}},
			{code: "refund", keywords: []string{"คืนเงิน", "refund", "cashback"}},
		}
	case TransactionTypeExpense:
		return []categoryRule{
			{code: "food", keywords: []string{"ข้าว", "อาหาร", "มื้อ", "ร้าน", "ก๋วยเตี๋ยว", "food", "lunch", "dinner"}},
			{code: "drink", keywords: []string{"กาแฟ", "ชา", "น้ำ", "เครื่องดื่ม", "coffee", "drink"}},
			{code: "transport", keywords: []string{"รถ", "แท็กซี่", "แทคซี่", "bts", "mrt", "grab", "น้ำมัน", "ทางด่วน", "transport"}},
			{code: "shopping", keywords: []string{"ซื้อ", "เสื้อ", "รองเท้า", "ของใช้", "shopping"}},
			{code: "housing", keywords: []string{"บ้าน", "คอนโด", "หอ", "ค่าเช่า", "rent"}},
			{code: "utilities", keywords: []string{"ไฟ", "น้ำประปา", "เน็ต", "โทรศัพท์", "internet", "ค่าไฟ", "ค่าน้ำ"}},
			{code: "health", keywords: []string{"ยา", "หมอ", "โรงพยาบาล", "สุขภาพ", "health"}},
			{code: "education", keywords: []string{"หนังสือ", "คอร์ส", "เรียน", "education"}},
			{code: "entertainment", keywords: []string{"หนัง", "netflix", "เกม", "concert", "บันเทิง"}},
			{code: "travel", keywords: []string{"โรงแรม", "ตั๋ว", "เที่ยว", "travel"}},
			{code: "family", keywords: []string{"พ่อ", "แม่", "ลูก", "ครอบครัว", "family"}},
			{code: "debt", keywords: []string{"ผ่อน", "หนี้", "บัตรเครดิต", "loan", "debt"}},
		}
	default:
		return nil
	}
}

func mustCategory(txType TransactionType, code, name string) Category {
	category, err := NewCategory(txType, code, name)
	if err != nil {
		panic(err)
	}
	return category
}
