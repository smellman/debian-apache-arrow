// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file_test

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/apache/arrow/go/v7/parquet"
	"github.com/apache/arrow/go/v7/parquet/file"
	"github.com/apache/arrow/go/v7/parquet/internal/testutils"
	"github.com/apache/arrow/go/v7/parquet/internal/utils"
	"github.com/apache/arrow/go/v7/parquet/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func initValues(values reflect.Value) {
	if values.Kind() != reflect.Slice {
		panic("must init values with slice")
	}

	r := rand.New(rand.NewSource(0))
	typ := values.Type().Elem()
	switch {
	case typ.Bits() <= 32:
		max := int64(math.MaxInt32)
		min := int64(math.MinInt32)
		for i := 0; i < values.Len(); i++ {
			values.Index(i).Set(reflect.ValueOf(r.Int63n(max-min+1) + min).Convert(reflect.TypeOf(int32(0))))
		}
	case typ.Bits() <= 64:
		max := int64(math.MaxInt64)
		min := int64(math.MinInt64)
		for i := 0; i < values.Len(); i++ {
			values.Index(i).Set(reflect.ValueOf(r.Int63n(max-min+1) + min))
		}
	}
}

func initDictValues(values reflect.Value, numDicts int) {
	repeatFactor := values.Len() / numDicts
	initValues(values)
	// add some repeated values
	for j := 1; j < repeatFactor; j++ {
		for i := 0; i < numDicts; i++ {
			values.Index(numDicts*j + i).Set(values.Index(i))
		}
	}
	// computed only dict_per_page * repeat_factor - 1 values < num_values compute remaining
	for i := numDicts * repeatFactor; i < values.Len(); i++ {
		values.Index(i).Set(values.Index(i - numDicts*repeatFactor))
	}
}

func makePages(version parquet.DataPageVersion, d *schema.Column, npages, lvlsPerPage int, typ reflect.Type, enc parquet.Encoding) ([]file.Page, int, reflect.Value, []int16, []int16) {
	nlevels := lvlsPerPage * npages
	nvalues := 0

	maxDef := d.MaxDefinitionLevel()
	maxRep := d.MaxRepetitionLevel()

	var (
		defLevels []int16
		repLevels []int16
	)

	valuesPerPage := make([]int, npages)
	if maxDef > 0 {
		defLevels = make([]int16, nlevels)
		testutils.FillRandomInt16(0, 0, maxDef, defLevels)
		for idx := range valuesPerPage {
			numPerPage := 0
			for i := 0; i < lvlsPerPage; i++ {
				if defLevels[i+idx*lvlsPerPage] == maxDef {
					numPerPage++
					nvalues++
				}
			}
			valuesPerPage[idx] = numPerPage
		}
	} else {
		nvalues = nlevels
		valuesPerPage[0] = lvlsPerPage
		for i := 1; i < len(valuesPerPage); i *= 2 {
			copy(valuesPerPage[i:], valuesPerPage[:i])
		}
	}

	if maxRep > 0 {
		repLevels = make([]int16, nlevels)
		testutils.FillRandomInt16(0, 0, maxRep, repLevels)
	}

	values := reflect.MakeSlice(reflect.SliceOf(typ), nvalues, nvalues)
	if enc == parquet.Encodings.Plain {
		initValues(values)
		return testutils.PaginatePlain(version, d, values, defLevels, repLevels, maxDef, maxRep, lvlsPerPage, valuesPerPage, parquet.Encodings.Plain), nvalues, values, defLevels, repLevels
	} else if enc == parquet.Encodings.PlainDict || enc == parquet.Encodings.RLEDict {
		initDictValues(values, lvlsPerPage)
		return testutils.PaginateDict(version, d, values, defLevels, repLevels, maxDef, maxRep, lvlsPerPage, valuesPerPage, parquet.Encodings.RLEDict), nvalues, values, defLevels, repLevels
	}
	panic("invalid encoding type for make pages")
}

func compareVectorWithDefLevels(left, right reflect.Value, defLevels []int16, maxDef, maxRep int16) assert.Comparison {
	return func() bool {
		if left.Kind() != reflect.Slice || right.Kind() != reflect.Slice {
			return false
		}

		if left.Type().Elem() != right.Type().Elem() {
			return false
		}

		iLeft, iRight := 0, 0
		for _, def := range defLevels {
			if def == maxDef {
				if !reflect.DeepEqual(left.Index(iLeft).Interface(), right.Index(iRight).Interface()) {
					return false
				}
				iLeft++
				iRight++
			} else if def == (maxDef - 1) {
				// null entry on the lowest nested level
				iRight++
			} else if def < (maxDef - 1) {
				// null entry on higher nesting level, only supported for non-repeating data
				if maxRep == 0 {
					iRight++
				}
			}
		}
		return true
	}
}

var mem = memory.DefaultAllocator

type PrimitiveReaderSuite struct {
	suite.Suite

	dataPageVersion parquet.DataPageVersion
	pager           file.PageReader
	reader          file.ColumnChunkReader
	pages           []file.Page
	values          reflect.Value
	defLevels       []int16
	repLevels       []int16
	nlevels         int
	nvalues         int
	maxDefLvl       int16
	maxRepLvl       int16
}

func (p *PrimitiveReaderSuite) TearDownTest() {
	p.clear()
}

func (p *PrimitiveReaderSuite) initReader(d *schema.Column) {
	m := new(testutils.MockPageReader)
	m.Test(p.T())
	m.TestData().Set("pages", p.pages)
	m.On("Err").Return((error)(nil))
	p.pager = m
	p.reader = file.NewColumnReader(d, m, mem)
}

func (p *PrimitiveReaderSuite) checkResults() {
	vresult := make([]int32, p.nvalues)
	dresult := make([]int16, p.nlevels)
	rresult := make([]int16, p.nlevels)

	var (
		read        int64 = 0
		totalRead   int   = 0
		batchActual int   = 0
		batchSize   int32 = 8
		batch       int   = 0
	)

	rdr := p.reader.(*file.Int32ColumnChunkReader)
	p.Require().NotNil(rdr)

	// this will cover both cases:
	// 1) batch size < page size (multiple ReadBatch from a single page)
	// 2) batch size > page size (BatchRead limits to single page)
	for {
		read, batch, _ = rdr.ReadBatch(int64(batchSize), vresult[totalRead:], dresult[batchActual:], rresult[batchActual:])
		totalRead += batch
		batchActual += int(read)
		batchSize = int32(utils.MinInt(1<<24, utils.MaxInt(int(batchSize*2), 4096)))
		if batch <= 0 {
			break
		}
	}

	p.Equal(p.nlevels, batchActual)
	p.Equal(p.nvalues, totalRead)
	p.Equal(p.values.Interface(), vresult)
	if p.maxDefLvl > 0 {
		p.Equal(p.defLevels, dresult)
	}
	if p.maxRepLvl > 0 {
		p.Equal(p.repLevels, rresult)
	}

	// catch improper writes at EOS
	read, batchActual, _ = rdr.ReadBatch(5, vresult, nil, nil)
	p.Zero(batchActual)
	p.Zero(read)
}

func (p *PrimitiveReaderSuite) clear() {
	p.values = reflect.ValueOf(nil)
	p.defLevels = nil
	p.repLevels = nil
	p.pages = nil
	p.pager = nil
	p.reader = nil
}

func (p *PrimitiveReaderSuite) testPlain(npages, levels int, d *schema.Column) {
	p.pages, p.nvalues, p.values, p.defLevels, p.repLevels = makePages(p.dataPageVersion, d, npages, levels, reflect.TypeOf(int32(0)), parquet.Encodings.Plain)
	p.nlevels = npages * levels
	p.initReader(d)
	p.checkResults()
	p.clear()
}

func (p *PrimitiveReaderSuite) testDict(npages, levels int, d *schema.Column) {
	p.pages, p.nvalues, p.values, p.defLevels, p.repLevels = makePages(p.dataPageVersion, d, npages, levels, reflect.TypeOf(int32(0)), parquet.Encodings.RLEDict)
	p.nlevels = npages * levels
	p.initReader(d)
	p.checkResults()
	p.clear()
}

func (p *PrimitiveReaderSuite) TestInt32FlatRequired() {
	const (
		levelsPerPage int = 100
		npages        int = 50
	)

	p.maxDefLvl = 0
	p.maxRepLvl = 0

	typ := schema.NewInt32Node("a", parquet.Repetitions.Required, -1)
	d := schema.NewColumn(typ, p.maxDefLvl, p.maxRepLvl)
	p.testPlain(npages, levelsPerPage, d)
	p.testDict(npages, levelsPerPage, d)
}

func (p *PrimitiveReaderSuite) TestInt32FlatOptional() {
	const (
		levelsPerPage int = 100
		npages        int = 50
	)

	p.maxDefLvl = 4
	p.maxRepLvl = 0
	typ := schema.NewInt32Node("b", parquet.Repetitions.Optional, -1)
	d := schema.NewColumn(typ, p.maxDefLvl, p.maxRepLvl)
	p.testPlain(npages, levelsPerPage, d)
	p.testDict(npages, levelsPerPage, d)
}

func (p *PrimitiveReaderSuite) TestInt32FlatRepeated() {
	const (
		levelsPerPage int = 100
		npages        int = 50
	)

	p.maxDefLvl = 4
	p.maxRepLvl = 2
	typ := schema.NewInt32Node("c", parquet.Repetitions.Repeated, -1)
	d := schema.NewColumn(typ, p.maxDefLvl, p.maxRepLvl)
	p.testPlain(npages, levelsPerPage, d)
	p.testDict(npages, levelsPerPage, d)
}

func (p *PrimitiveReaderSuite) TestReadBatchMultiPage() {
	const (
		levelsPerPage int = 100
		npages        int = 3
	)

	p.maxDefLvl = 0
	p.maxRepLvl = 0
	typ := schema.NewInt32Node("a", parquet.Repetitions.Required, -1)
	d := schema.NewColumn(typ, p.maxDefLvl, p.maxRepLvl)
	p.pages, p.nvalues, p.values, p.defLevels, p.repLevels = makePages(p.dataPageVersion, d, npages, levelsPerPage, reflect.TypeOf(int32(0)), parquet.Encodings.Plain)
	p.initReader(d)

	vresult := make([]int32, levelsPerPage*npages)
	dresult := make([]int16, levelsPerPage*npages)
	rresult := make([]int16, levelsPerPage*npages)

	rdr := p.reader.(*file.Int32ColumnChunkReader)
	total, read, err := rdr.ReadBatch(int64(levelsPerPage*npages), vresult, dresult, rresult)
	p.NoError(err)
	p.EqualValues(levelsPerPage*npages, total)
	p.EqualValues(levelsPerPage*npages, read)
}

func (p *PrimitiveReaderSuite) TestInt32FlatRequiredSkip() {
	const (
		levelsPerPage int = 100
		npages        int = 5
	)

	p.maxDefLvl = 0
	p.maxRepLvl = 0
	typ := schema.NewInt32Node("a", parquet.Repetitions.Required, -1)
	d := schema.NewColumn(typ, p.maxDefLvl, p.maxRepLvl)
	p.pages, p.nvalues, p.values, p.defLevels, p.repLevels = makePages(p.dataPageVersion, d, npages, levelsPerPage, reflect.TypeOf(int32(0)), parquet.Encodings.Plain)
	p.initReader(d)

	vresult := make([]int32, levelsPerPage/2)
	dresult := make([]int16, levelsPerPage/2)
	rresult := make([]int16, levelsPerPage/2)

	rdr := p.reader.(*file.Int32ColumnChunkReader)

	p.Run("skip_size > page_size", func() {
		// Skip first 2 pages
		skipped, _ := rdr.Skip(int64(2 * levelsPerPage))
		p.Equal(int64(2*levelsPerPage), skipped)

		rdr.ReadBatch(int64(levelsPerPage/2), vresult, dresult, rresult)
		subVals := p.values.Slice(2*levelsPerPage, int(2.5*float64(levelsPerPage))).Interface().([]int32)
		p.Equal(subVals, vresult)
	})

	p.Run("skip_size == page_size", func() {
		// skip across two pages
		skipped, _ := rdr.Skip(int64(levelsPerPage))
		p.Equal(int64(levelsPerPage), skipped)
		// read half a page
		rdr.ReadBatch(int64(levelsPerPage/2), vresult, dresult, rresult)
		subVals := p.values.Slice(int(3.5*float64(levelsPerPage)), 4*levelsPerPage).Interface().([]int32)
		p.Equal(subVals, vresult)
	})

	p.Run("skip_size < page_size", func() {
		// skip limited to a single page
		// Skip half a page
		skipped, _ := rdr.Skip(int64(levelsPerPage / 2))
		p.Equal(int64(0.5*float32(levelsPerPage)), skipped)
		// Read half a page
		rdr.ReadBatch(int64(levelsPerPage/2), vresult, dresult, rresult)
		subVals := p.values.Slice(int(4.5*float64(levelsPerPage)), p.values.Len()).Interface().([]int32)
		p.Equal(subVals, vresult)
	})
}

func (p *PrimitiveReaderSuite) TestDictionaryEncodedPages() {
	p.maxDefLvl = 0
	p.maxRepLvl = 0
	typ := schema.NewInt32Node("a", parquet.Repetitions.Required, -1)
	descr := schema.NewColumn(typ, p.maxDefLvl, p.maxRepLvl)
	dummy := memory.NewResizableBuffer(mem)

	p.Run("Dict: Plain, Data: RLEDict", func() {
		dictPage := file.NewDictionaryPage(dummy, 0, parquet.Encodings.Plain)
		dataPage := testutils.MakeDataPage(p.dataPageVersion, descr, nil, 0, parquet.Encodings.RLEDict, dummy, nil, nil, 0, 0)

		p.pages = append(p.pages, dictPage, dataPage)
		p.initReader(descr)
		p.NotPanics(func() { p.reader.HasNext() })
		p.NoError(p.reader.Err())
		p.pages = p.pages[:0]
	})

	p.Run("Dict: Plain Dictionary, Data: Plain Dictionary", func() {
		dictPage := file.NewDictionaryPage(dummy, 0, parquet.Encodings.PlainDict)
		dataPage := testutils.MakeDataPage(p.dataPageVersion, descr, nil, 0, parquet.Encodings.PlainDict, dummy, nil, nil, 0, 0)
		p.pages = append(p.pages, dictPage, dataPage)
		p.initReader(descr)
		p.NotPanics(func() { p.reader.HasNext() })
		p.NoError(p.reader.Err())
		p.pages = p.pages[:0]
	})

	p.Run("Panic if dict page not first", func() {
		dataPage := testutils.MakeDataPage(p.dataPageVersion, descr, nil, 0, parquet.Encodings.RLEDict, dummy, nil, nil, 0, 0)
		p.pages = append(p.pages, dataPage)
		p.initReader(descr)
		p.NotPanics(func() { p.False(p.reader.HasNext()) })
		p.Error(p.reader.Err())
		p.pages = p.pages[:0]
	})

	p.Run("Only RLE is supported", func() {
		dictPage := file.NewDictionaryPage(dummy, 0, parquet.Encodings.DeltaByteArray)
		p.pages = append(p.pages, dictPage)
		p.initReader(descr)
		p.NotPanics(func() { p.False(p.reader.HasNext()) })
		p.Error(p.reader.Err())
		p.pages = p.pages[:0]
	})

	p.Run("Cannot have more than one dict", func() {
		dictPage1 := file.NewDictionaryPage(dummy, 0, parquet.Encodings.PlainDict)
		dictPage2 := file.NewDictionaryPage(dummy, 0, parquet.Encodings.Plain)
		p.pages = append(p.pages, dictPage1, dictPage2)
		p.initReader(descr)
		p.NotPanics(func() { p.False(p.reader.HasNext()) })
		p.Error(p.reader.Err())
		p.pages = p.pages[:0]
	})

	p.Run("Unsupported encoding", func() {
		dataPage := testutils.MakeDataPage(p.dataPageVersion, descr, nil, 0, parquet.Encodings.DeltaByteArray, dummy, nil, nil, 0, 0)
		p.pages = append(p.pages, dataPage)
		p.initReader(descr)
		p.Panics(func() { p.reader.HasNext() })
		// p.Error(p.reader.Err())
		p.pages = p.pages[:0]
	})

	p.pages = p.pages[:2]
}

func TestPrimitiveReader(t *testing.T) {
	t.Parallel()
	t.Run("datapage v1", func(t *testing.T) {
		suite.Run(t, new(PrimitiveReaderSuite))
	})
	t.Run("datapage v2", func(t *testing.T) {
		suite.Run(t, &PrimitiveReaderSuite{dataPageVersion: parquet.DataPageV2})
	})
}
