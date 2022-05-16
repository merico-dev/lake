package helper

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Insert data by batch can increase database performance drastically, this class aim to make batch-save easier,
// It takes care the database operation for specified `slotType`, records got saved into database whenever cache hits
// The `size` limit, remember to call the `Close` method to save the last batch
type BatchSave struct {
	slotType reflect.Type
	// slots can not be []interface{}, because gorm wouldn't take it
	// I'm guessing the reason is the type information lost when converted to interface{}
	slots      reflect.Value
	db         *gorm.DB
	current    int
	size       int
	valueIndex map[string]int
}

func NewBatchSave(db *gorm.DB, slotType reflect.Type, size int) (*BatchSave, error) {
	if slotType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("slotType must be a pointer")
	}
	if !hasPrimaryKey(slotType) {
		return nil, fmt.Errorf("no primary key")
	}
	return &BatchSave{
		slotType:   slotType,
		slots:      reflect.MakeSlice(reflect.SliceOf(slotType), size, size),
		db:         db,
		size:       size,
		valueIndex: make(map[string]int),
	}, nil
}

func (c *BatchSave) Add(slot interface{}) error {
	// type checking
	if reflect.TypeOf(slot) != c.slotType {
		return fmt.Errorf("sub cache type mismatched")
	}
	if reflect.ValueOf(slot).Kind() != reflect.Ptr {
		return fmt.Errorf("slot is not a pointer")
	}
	// deduplication
	key := getPrimaryKeyValue(slot)

	if key != "" {
		if index, ok := c.valueIndex[key]; !ok {
			c.valueIndex[key] = c.current
		} else {
			c.slots.Index(index).Set(reflect.ValueOf(slot))
			return nil
		}
	}
	c.slots.Index(c.current).Set(reflect.ValueOf(slot))
	c.current++
	// flush out into database if max outed
	if c.current == c.size {
		return c.Flush()
	}
	return nil
}

func (c *BatchSave) Flush() error {
	err := c.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(c.slots.Slice(0, c.current).Interface()).Error
	if err != nil {
		return err
	}
	c.current = 0
	c.valueIndex = make(map[string]int)
	return nil
}

func (c *BatchSave) Close() error {
	if c.current > 0 {
		return c.Flush()
	}
	return nil
}

func isPrimaryKey(f reflect.StructField) bool {
	tag := strings.TrimSpace(f.Tag.Get("gorm"))
	return strings.HasPrefix(strings.ToLower(tag), "primarykey")
}

func hasPrimaryKey(ifv reflect.Type) bool {
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		v := ifv.Field(i)
		switch v.Type.Kind() {
		case reflect.Struct:
			ok := hasPrimaryKey(v.Type)
			if ok {
				return true
			}
		default:
			if ok := isPrimaryKey(v); ok {
				return true
			}
		}
	}
	return false
}

func getPrimaryKeyValue(iface interface{}) string {
	var ss []string
	ifv := reflect.ValueOf(iface)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		v := ifv.Field(i)
		switch v.Kind() {
		case reflect.Struct:
			s := getPrimaryKeyValue(v.Interface())
			if s != "" {
				ss = append(ss, s)
			}
		default:
			if isPrimaryKey(ifv.Type().Field(i)) {
				s := fmt.Sprintf("%v", v.Interface())
				if s != "" {
					ss = append(ss, s)
				}
			}
		}
	}
	return strings.Join(ss, ":")
}
