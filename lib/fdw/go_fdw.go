package main

//#cgo CFLAGS: -Iinclude/postgresql/server -Iinclude/postgresql/internal
//#cgo LDFLAGS: -Wl,-unresolved-symbols=ignore-all
//
//#include "postgres.h"
//#include "funcapi.h"
//#include "fmgr.h"
//#include "access/relscan.h"
//#include "catalog/pg_type.h"
//#include "commands/defrem.h"
//#include "foreign/fdwapi.h"
//#include "foreign/foreign.h"
//#include "nodes/extensible.h"
//#include "optimizer/pathnode.h"
//#include "utils/builtins.h"
//#include "utils/elog.h"
//#include "utils/fmgrprotos.h"
//#include "utils/tuplestore.h"
//#include "utils/jsonb.h"
//
//typedef struct GoFdwExecutionState
//{
// uint tok;
//} GoFdwExecutionState;
//
//static inline GoFdwExecutionState* makeState(){
//  GoFdwExecutionState *s = (GoFdwExecutionState *) malloc(sizeof(GoFdwExecutionState));
//  return s;
//}
//
//static inline DefElem* cellGetDef(ListCell *n) { return (DefElem*)n->data.ptr_value; }
//
//static inline void freeState(GoFdwExecutionState * s){ if (s) free(s); }
//static void eReport(const char *msg) {
//  ereport(ERROR, (errcode(ERRCODE_FDW_ERROR), errmsg("%s", msg)));
//}
//
//static Datum cStringGetDatum(const char *str) {
//  PG_RETURN_TEXT_P(CStringGetTextDatum(str));
//}
//
//static Datum jsonGetDatum(const char *str) {
//  PG_RETURN_JSONB(DirectFunctionCall1(jsonb_in, CStringGetDatum(str)));
//}
//
//static Datum numericGetDatum(int64_t num) {
//  PG_RETURN_INT64(Int64GetDatum(num));
//}
//
//static void saveTuple(Datum *data, bool *isnull, ScanState *state) {
//  HeapTuple tuple = heap_form_tuple(state->ss_currentRelation->rd_att, data, isnull);
//  ExecStoreTuple(tuple, state->ss_ScanTupleSlot, InvalidBuffer, false);
//}
import "C"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

var table Table

// SetTable sets a Go objects that will receive all FDW requests.
// Should be called with user-defined implementation in init().
func SetTable(t Table) { table = t }

// Explainer is an helper build an EXPLAIN response.
type Explainer struct {
	es *C.ExplainState
}

// Property adds a key-value property to results of EXPLAIN query.
func (e Explainer) Property(k, v string) {
	C.ExplainPropertyText(C.CString(k), C.CString(v), e.es)
}

// Options is a set of FDW options provided by user during table creation.
type Options struct {
	TableOptions  map[string]string
	ServerOptions map[string]string
}

// Table is a main interface for FDW table.
//
// If there are multiple tables created with this module, they can be identified by table options.
type Table interface {
	// Stats returns stats for a table.
	Stats(opts *Options) (TableStats, error)
	// Scan starts a new scan of the table.
	// Iterator should not load data instantly, since Scan will be called for EXPLAIN as well.
	// Results should be fetched during Next calls.
	Scan(rel *Relation, opts *Options) (Iterator, error)
}

// Iterator is an interface for table scanner implementations.
type Iterator interface {
	// Next returns next row (tuple). Nil slice means there is no more rows to scan.
	Next() ([]interface{}, error)
	// Reset restarts an iterator from the beginning (possible with a new data snapshot).
	Reset()
	// Close stops an iteration and frees any resources.
	Close() error
}

// Explainable is an optional interface for Iterator that can explain it's execution plan.
type Explainable interface {
	// Explain is called during EXPLAIN query.
	Explain(e Explainer)
}

type Relation struct {
	ID      Oid
	IsValid bool
	Attr    *TupleDesc
}

type TupleDesc struct {
	Raw     C.TupleDesc
	TypeID  Oid
	TypeMod int
	HasOid  bool
	Attrs   []Attr // columns
}

type Attr struct {
	Name    string
	Type    Oid
	NotNull bool
	Options map[string]string
}

type TableStats struct {
	Rows      uint // an estimated number of rows
	StartCost Cost
	TotalCost Cost
}

// Cost is a approximate cost of an operation. See Postgres docs for details.
type Cost float64

// Oid is a Postgres internal object ID.
type Oid uint

// A list of constants for Postgres data types
const (
	TypeBool        = Oid(C.BOOLOID)
	TypeBytes       = Oid(C.BYTEAOID)
	TypeChar        = Oid(C.CHAROID)
	TypeName        = Oid(C.NAMEOID)
	TypeInt64       = Oid(C.INT8OID)
	TypeInt16       = Oid(C.INT2OID)
	TypeInt16Vector = Oid(C.INT2VECTOROID)
	TypeInt32       = Oid(C.INT4OID)
	TypeRegProc     = Oid(C.REGPROCOID)
	TypeText        = Oid(C.TEXTOID)
	TypeOid         = Oid(C.OIDOID)

	TypeJson = Oid(C.JSONOID)
	TypeXml  = Oid(C.XMLOID)

	TypeFloat32 = Oid(C.FLOAT4OID)
	TypeFloat64 = Oid(C.FLOAT8OID)

	TypeTimestamp = Oid(C.TIMESTAMPOID)
	TypeInterval  = Oid(C.INTERVALOID)
)

// FIXME: some black magic here; we save pointers to all necessary functions (passed by glue C code)
// FIXME: it would be better to link to pg executable properly

var (
	fmu sync.Mutex
)

//export goAnalyzeForeignTable
func goAnalyzeForeignTable(relation C.Relation, fnc *C.AcquireSampleRowsFunc, totalpages *C.BlockNumber) C.bool {
	*totalpages = 1
	return 1
}

//export goGetForeignRelSize
func goGetForeignRelSize(root *C.PlannerInfo, baserel *C.RelOptInfo, foreigntableid C.Oid) {
	// Obtain relation size estimates for a foreign table
	opts := getFTableOptions(Oid(foreigntableid))
	st, err := table.Stats(opts)
	if err != nil {
		errReport(err)
		return
	}
	baserel.rows = C.double(st.Rows)
	baserel.fdw_private = nil
}

//export goExplainForeignScan
func goExplainForeignScan(node *C.ForeignScanState, es *C.ExplainState) {
	s := getState(node.fdw_state)
	if s == nil {
		return
	}
	// Produce extra output for EXPLAIN
	if e, ok := s.Iter.(Explainable); ok {
		e.Explain(Explainer{es: es})
	}

	cs := (*C.GoFdwExecutionState)(node.fdw_state)
	clearState(uint64(cs.tok))
	C.freeState(cs)
	node.fdw_state = nil
}

//export goGetForeignPaths
func goGetForeignPaths(root *C.PlannerInfo, baserel *C.RelOptInfo, foreigntableid C.Oid) {
	// Create possible access paths for a scan on the foreign table
	opts := getFTableOptions(Oid(foreigntableid))
	st, err := table.Stats(opts)
	if err != nil {
		errReport(err)
		return
	}
	C.add_path(baserel,
		(*C.Path)(unsafe.Pointer(C.create_foreignscan_path(
			root,
			baserel,
			baserel.reltarget,
			baserel.rows,
			C.Cost(st.StartCost),
			C.Cost(st.TotalCost),
			nil, // no pathkeys
			nil, // no outer rel either
			nil, // no extra plan
			nil,
		))),
	)
}

//export goBeginForeignScan
func goBeginForeignScan(node *C.ForeignScanState, eflags C.int) {
	rel := buildRelation(node.ss.ss_currentRelation)
	opts := getFTableOptions(rel.ID)
	iter, err := table.Scan(rel, opts)
	if err != nil {
		errReport(err)
		return
	}
	s := &state{
		Rel:  rel,
		Opts: opts,
		Iter: iter,
	}
	i := saveState(s)

	//plan := node.ss.ps.plan
	//plan := (*C.ForeignScan)(unsafe.Pointer(node.ss.ps.plan))

	cs := C.makeState()
	cs.tok = C.uint(i)
	node.fdw_state = unsafe.Pointer(cs)

	//if eflags&C.EXEC_FLAG_EXPLAIN_ONLY != 0 {
	//	return // Do nothing in EXPLAIN
	//}
}

//export goIterateForeignScan
func goIterateForeignScan(node *C.ForeignScanState) *C.TupleTableSlot {
	s := getState(node.fdw_state)

	slot := node.ss.ss_ScanTupleSlot
	C.ExecClearTuple(slot)

	row, err := s.Iter.Next()
	if err != nil {
		errReport(err)
		return nil
	}
	if row == nil {
		return slot
	}

	isNull := make([]C.bool, len(row))
	data := make([]C.Datum, len(row))

	for i, v := range row {
		if v == nil {
			isNull[i] = C.bool(1)
			continue
		}

		datum, err := valToDatum(v)
		if err != nil {
			errReport(err)
			return nil
		}
		// everyone loves manually calculating array offsets
		data[i] = datum
	}

	C.saveTuple(&data[0], &isNull[0], &node.ss)
	return slot
}

//export goReScanForeignScan
func goReScanForeignScan(node *C.ForeignScanState) {
	// Rescan table, possibly with new parameters
	s := getState(node.fdw_state)
	s.Iter.Reset()
}

//export goEndForeignScan
func goEndForeignScan(node *C.ForeignScanState) {
	// Finish scanning foreign table and dispose objects used for this scan
	s := getState(node.fdw_state)
	if s == nil {
		return
	}
	cs := (*C.GoFdwExecutionState)(node.fdw_state)
	clearState(uint64(cs.tok))
	C.freeState(cs)
	node.fdw_state = nil
}

type state struct {
	Rel  *Relation
	Opts *Options
	Iter Iterator
}

var (
	mu   sync.RWMutex
	si   uint64
	sess = make(map[uint64]*state)
)

func saveState(s *state) uint64 {
	mu.Lock()
	si++
	i := si
	sess[i] = s
	mu.Unlock()
	return i
}

func clearState(i uint64) {
	mu.Lock()
	delete(sess, i)
	mu.Unlock()
}

func getState(p unsafe.Pointer) *state {
	if p == nil {
		return nil
	}
	cs := (*C.GoFdwExecutionState)(p)
	i := uint64(cs.tok)
	mu.RLock()
	s := sess[i]
	mu.RUnlock()
	return s
}

func getFTableOptions(id Oid) *Options {
	f := C.GetForeignTable(C.Oid(id))
	s := C.GetForeignServer(C.Oid(f.serverid))
	return &Options{
		TableOptions:  getOptions(f.options),
		ServerOptions: getOptions(s.options),
	}
}

func getOptions(opts *C.List) map[string]string {
	m := make(map[string]string)
	if opts == nil {
		return m
	}
	for it := opts.head; it != nil; it = it.next {
		el := C.cellGetDef(it)
		name := C.GoString(el.defname)
		val := C.GoString(C.defGetString(el))
		m[name] = val
	}
	return m
}

func buildRelation(rel C.Relation) *Relation {
	r := &Relation{
		ID:      Oid(rel.rd_id),
		IsValid: goBool(rel.rd_isvalid),
		Attr:    buildTupleDesc(rel.rd_att),
	}
	return r
}

func goBool(b C.bool) bool {
	return b != 0
}

func goString(p unsafe.Pointer, n int) string {
	b := C.GoBytes(p, C.int(n))
	i := bytes.IndexByte(b, 0)
	if i < 0 {
		i = len(b)
	}
	return string(b[:i])
}

func buildTupleDesc(desc C.TupleDesc) *TupleDesc {
	if desc == nil {
		return nil
	}
	d := &TupleDesc{
		TypeID:  Oid(desc.tdtypeid),
		TypeMod: int(desc.tdtypmod),
		HasOid:  goBool(desc.tdhasoid),
		Attrs:   make([]Attr, 0, int(desc.natts)),
	}
	for i := 0; i < cap(d.Attrs); i++ {
		off := uintptr(i) * uintptr(C.sizeof_Form_pg_attribute)
		p := *(*C.Form_pg_attribute)(unsafe.Pointer(uintptr(unsafe.Pointer(desc.attrs)) + off))
		d.Attrs = append(d.Attrs, buildAttr(p))
	}
	return d
}

const nameLen = C.NAMEDATALEN

func buildAttr(attr *C.FormData_pg_attribute) (out Attr) {
	out.Name = goString(unsafe.Pointer(&attr.attname.data[0]), nameLen)
	out.Type = Oid(attr.atttypid)
	out.NotNull = goBool(attr.attnotnull)
	out.Options = getOptions(C.GetForeignColumnOptions(attr.attrelid, attr.attnum))

	return
}

func errReport(err error) {
	C.eReport(C.CString(err.Error()))
}

func valToDatum(v interface{}) (C.Datum, error) {
	// TODO(EKF): handle more datatypes
	// TODO(EKF): take column type into account
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := reflect.ValueOf(v).Int()
		return C.numericGetDatum(C.int64_t(val)), nil
	case reflect.Map:
		bytes, err := json.Marshal(v)
		if err != nil {
			return C.Datum(0), errors.Wrapf(err, "couldn't marshall %v", v)
		}
		value := C.CString(string(bytes))
		return C.jsonGetDatum(value), nil
	case reflect.String:
		value := C.CString(fmt.Sprintf("%s", v))
		return C.cStringGetDatum(value), nil
	default:
		value := C.CString(fmt.Sprintf("%v", v))
		return C.cStringGetDatum(value), nil
	}

}

// required by buildmode=c-archive
func main() {}
