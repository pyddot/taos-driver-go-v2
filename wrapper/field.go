package wrapper

/*
#include <taos.h>
*/
import "C"
import (
	"bytes"
	"reflect"
	"unsafe"

	"github.com/pyddot/taos-driver-go-v2/common"
	"github.com/pyddot/taos-driver-go-v2/errors"
)

type RowsHeader struct {
	ColNames  []string
	ColTypes  []uint8
	ColLength []uint16
}

func ReadColumn(result unsafe.Pointer, count int) (*RowsHeader, error) {
	if result == nil {
		return nil, &errors.TaosError{Code: 0xffff, ErrStr: "invalid result"}
	}
	rowsHeader := &RowsHeader{
		ColNames:  make([]string, count),
		ColTypes:  make([]uint8, count),
		ColLength: make([]uint16, count),
	}
	pFields := TaosFetchFields(result)
	for i := 0; i < count; i++ {
		field := *(*C.struct_taosField)(unsafe.Pointer(uintptr(pFields) + uintptr(C.sizeof_struct_taosField*C.int(i))))
		buf := bytes.NewBufferString("")
		for _, c := range field.name {
			if c == 0 {
				break
			}
			buf.WriteByte(byte(c))
		}
		rowsHeader.ColNames[i] = buf.String()
		rowsHeader.ColTypes[i] = (uint8)(field._type)
		rowsHeader.ColLength[i] = (uint16)(field.bytes)
	}
	return rowsHeader, nil
}

func (rh *RowsHeader) TypeDatabaseName(i int) string {
	return common.TypeNameMap[int(rh.ColTypes[i])]
}

func (rh *RowsHeader) ScanType(i int) reflect.Type {
	t, exist := common.ColumnTypeMap[int(rh.ColTypes[i])]
	if !exist {
		return common.UnknownType
	}
	return t
}

func FetchLengths(res unsafe.Pointer, count int) []int {
	lengths := TaosFetchLengths(res)
	result := make([]int, count)
	for i := 0; i < count; i++ {
		result[i] = int(*(*C.int)(unsafe.Pointer(uintptr(lengths) + uintptr(C.sizeof_int*C.int(i)))))
	}
	return result
}
