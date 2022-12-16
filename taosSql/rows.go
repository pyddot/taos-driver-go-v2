package taosSql

import (
	"database/sql/driver"
	"io"
	"reflect"
	"unsafe"

	"github.com/pyddot/taos-driver-go-v2/errors"
	"github.com/pyddot/taos-driver-go-v2/wrapper"
	"github.com/pyddot/taos-driver-go-v2/wrapper/handler"
)

type rows struct {
	handler     *handler.Handler
	rowsHeader  *wrapper.RowsHeader
	done        bool
	block       unsafe.Pointer
	blockOffset int
	blockSize   int
	lengthList  []int
	result      unsafe.Pointer
}

func (rs *rows) Columns() []string {
	return rs.rowsHeader.ColNames
}

func (rs *rows) ColumnTypeDatabaseTypeName(i int) string {
	return rs.rowsHeader.TypeDatabaseName(i)
}

func (rs *rows) ColumnTypeLength(i int) (length int64, ok bool) {
	return int64(rs.rowsHeader.ColLength[i]), true
}

func (rs *rows) ColumnTypeScanType(i int) reflect.Type {
	return rs.rowsHeader.ScanType(i)
}

func (rs *rows) Close() error {
	rs.freeResult()
	rs.block = nil
	return nil
}

func (rs *rows) Next(dest []driver.Value) error {
	if rs.done {
		return io.EOF
	}

	if rs.result == nil {
		return &errors.TaosError{Code: 0xffff, ErrStr: "result is nil!"}
	}

	if rs.block == nil {
		err := rs.taosFetchBlock()
		if err != nil {
			return err
		}
	}
	if rs.blockSize == 0 {
		rs.block = nil
		return io.EOF
	}

	if rs.blockOffset >= rs.blockSize {
		err := rs.taosFetchBlock()
		if err != nil {
			return err
		}
	}
	if rs.blockSize == 0 {
		rs.block = nil
		return io.EOF
	}
	wrapper.ReadRow(dest, rs.result, rs.block, rs.blockOffset, rs.lengthList, rs.rowsHeader.ColTypes)
	rs.blockOffset++
	return nil
}

func (rs *rows) taosFetchBlock() error {
	result := rs.asyncFetchRows()
	if result.N == 0 {
		rs.blockSize = 0
		return nil
	} else {
		if result.N < 0 {
			code := wrapper.TaosError(result.Res)
			errStr := wrapper.TaosErrorStr(result.Res)
			return errors.NewError(code, errStr)
		}
	}
	rs.blockSize = result.N
	rs.block = wrapper.TaosResultBlock(result.Res)
	rs.lengthList = wrapper.FetchLengths(rs.result, len(rs.rowsHeader.ColLength))
	rs.blockOffset = 0
	return nil
}

func (rs *rows) asyncFetchRows() *handler.AsyncResult {
	locker.Lock()
	wrapper.TaosFetchRowsA(rs.result, rs.handler.Handler)
	locker.Unlock()
	r := <-rs.handler.Caller.FetchResult
	return r
}

func (rs *rows) freeResult() {
	asyncHandlerPool.Put(rs.handler)
	if rs.result != nil {
		locker.Lock()
		wrapper.TaosFreeResult(rs.result)
		locker.Unlock()
		rs.result = nil
	}
}
