package irx

import (
	"fmt"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/value"
)

// A fixer keeps track of global and local identifiers to replace dummy values
// with their real values.
type fixer struct {
	// globals maps global identifiers to their real values.
	globals map[string]value.Value
	// locals maps local identifiers to their real values.
	locals map[string]value.Value
}

// getGlobal returns the global value of the given global identifier.
func (fix *fixer) getGlobal(name string) value.Value {
	global, ok := fix.globals[name]
	if !ok {
		panic(fmt.Sprintf("unable to locate global identifier %q", name))
	}
	return global
}

// getFunc returns the function of the given function name.
func (fix *fixer) getFunc(name string) *ir.Function {
	global := fix.getGlobal(name)
	f, ok := global.(*ir.Function)
	if !ok {
		panic(fmt.Sprintf("invalid function type; expected *ir.Function, got %T", global))
	}
	return f
}

// getLocal returns the local value of the given local identifier.
func (fix *fixer) getLocal(name string) value.Value {
	local, ok := fix.locals[name]
	if !ok {
		panic(fmt.Sprintf("unable to locate local identifier %q", name))
	}
	return local
}

// getBlock returns the basic block of the given label name.
func (fix *fixer) getBloack(name string) *ir.BasicBlock {
	local := fix.getLocal(name)
	block, ok := local.(*ir.BasicBlock)
	if !ok {
		panic(fmt.Sprintf("invalid basic block type; expected *ir.BasicBlock, got %T", local))
	}
	return block
}

// === [ Modules ] =============================================================

// fixModule replaces dummy values within the given module with their real
// values.
func fixModule(m *ir.Module) *ir.Module {
	fix := &fixer{
		globals: make(map[string]value.Value),
	}

	// Index global variables.
	for _, global := range m.Globals() {
		name := global.Name()
		if _, ok := fix.globals[name]; ok {
			panic(fmt.Sprintf("global identifier %q already present; old `%v`, new `%v`", name, fix.globals[name], global))
		}
		fix.globals[name] = global
	}

	// TODO: Remove debug output.
	//fmt.Println("=== [ globals ] ===")
	//pretty.Println(fix.globals)

	// Index functions.
	for _, f := range m.Funcs() {
		name := f.Name()
		if _, ok := fix.globals[name]; ok {
			panic(fmt.Sprintf("global identifier %q already present; old `%v`, new `%v`", name, fix.globals[name], f))
		}
		fix.globals[name] = f
	}

	// TODO: Figure out if global variables ever may contain dummy values.

	// Fix functions.
	for _, f := range m.Funcs() {
		fix.fixFunction(f)
	}

	return m
}

// === [ Functions ] ===========================================================

// fixFunction replaces dummy values within the given function with their real
// values.
func (fix *fixer) fixFunction(f *ir.Function) {
	// Reset locals.
	fix.locals = make(map[string]value.Value)

	// Force generate local IDs.
	_ = f.String()

	// Index basic blocks.
	blocks := f.Blocks()
	for _, block := range blocks {
		name := block.Name()
		if _, ok := fix.locals[name]; ok {
			panic(fmt.Sprintf("basic block label %q already present; old `%v`, new `%v`", name, fix.locals[name], block))
		}
		fix.locals[name] = block

		// Index instructions producing local variables.
		var insts []ir.Instruction
		for _, inst := range block.Insts() {
			// Fix dummy instructions before indexing local variables.
			switch old := inst.(type) {
			case *instCallDummy:
				inst = fix.fixCallInstDummy(old)
			}
			insts = append(insts, inst)
			if inst, ok := inst.(value.Value); ok {
				name := stripLocal(inst.Ident())
				if _, ok := fix.locals[name]; ok {
					panic(fmt.Sprintf("instruction name %q already present; old `%v`, new `%v`", name, fix.locals[name], inst))
				}
				fix.locals[name] = inst
			}
		}
		block.SetInsts(insts)
	}

	// Fix dummy terminators.
	for _, block := range blocks {
		term := block.Term()
		switch old := term.(type) {
		case *termBrDummy:
			term = fix.fixBrTermDummy(old)
		case *termCondBrDummy:
			term = fix.fixCondBrTermDummy(old)
		}
		block.SetTerm(term)
	}

	// TODO: Remove debug output.
	//fmt.Printf("=== [ locals of %q ] ===\n", f.Name())
	//pretty.Println(fix.locals)

	// Fix basic blocks.
	for _, block := range blocks {
		fix.fixBlock(block)
	}
}

// === [ Values ] ==============================================================

// fixValue replaces given dummy value with its real value. The boolean return
// value indicates if a dummy value was replaced.
func (fix *fixer) fixValue(old value.Value) (value.Value, bool) {
	// TODO: Add all instructions producing values.
	switch old := old.(type) {
	case *globalDummy:
		return fix.getGlobal(old.name), true
	case *localDummy:
		return fix.getLocal(old.name), true
	case *constant.Int:
		// nothing to do; valid value.
	case *constant.Float:
		// nothing to do; valid value.
	case *ir.InstAdd, *ir.InstFAdd, *ir.InstSub, *ir.InstFSub, *ir.InstMul, *ir.InstFMul, *ir.InstUDiv, *ir.InstSDiv, *ir.InstFDiv, *ir.InstURem, *ir.InstSRem, *ir.InstFRem:
		// nothing to do; valid value.
	case *ir.InstICmp:
		// nothing to do; valid value.
	default:
		panic(fmt.Sprintf("support for value type %T not yet implemented", old))
	}
	return old, false
}

// === [ Basic blocks ] ========================================================

// fixBlock replaces dummy values within the given basic block with their real
// values.
func (fix *fixer) fixBlock(block *ir.BasicBlock) {
	// Fix instructions.
	var insts []ir.Instruction
	for _, inst := range block.Insts() {
		inst = fix.fixInst(inst)
		insts = append(insts, inst)
	}
	block.SetInsts(insts)

	// Fix terminator.
	term := fix.fixTerm(block.Term())
	block.SetTerm(term)
}

// === [ Instructions ] ========================================================

// fixInst replaces dummy values within the given instruction with their real
// values.
func (fix *fixer) fixInst(inst ir.Instruction) ir.Instruction {
	switch inst := inst.(type) {
	// Binary instructions
	case *ir.InstAdd:
		return fix.fixAddInst(inst)
	case *ir.InstFAdd:
		return fix.fixFAddInst(inst)
	case *ir.InstSub:
		return fix.fixSubInst(inst)
	case *ir.InstFSub:
		return fix.fixFSubInst(inst)
	case *ir.InstMul:
		return fix.fixMulInst(inst)
	case *ir.InstFMul:
		return fix.fixFMulInst(inst)
	case *ir.InstUDiv:
		return fix.fixUDivInst(inst)
	case *ir.InstSDiv:
		return fix.fixSDivInst(inst)
	case *ir.InstFDiv:
		return fix.fixFDivInst(inst)
	case *ir.InstURem:
		return fix.fixURemInst(inst)
	case *ir.InstSRem:
		return fix.fixSRemInst(inst)
	case *ir.InstFRem:
		return fix.fixFRemInst(inst)
	// Bitwise instructions
	// Memory instructions
	case *ir.InstLoad:
		return fix.fixLoadInst(inst)
	case *ir.InstStore:
		return fix.fixStoreInst(inst)
	// Conversion instructions
	// Other instructions
	case *ir.InstICmp:
		return fix.fixICmpInst(inst)
	case *ir.InstCall:
		return fix.fixCallInst(inst)
	default:
		panic(fmt.Sprintf("support for instruction type %T not yet implemented", inst))
	}
}

// --- [ Binary instructions ] -------------------------------------------------

// fixAddInst replaces dummy values within the given add instruction with their
// real values.
func (fix *fixer) fixAddInst(old *ir.InstAdd) *ir.InstAdd {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixFAddInst replaces dummy values within the given fadd instruction with their
// real values.
func (fix *fixer) fixFAddInst(old *ir.InstFAdd) *ir.InstFAdd {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixSubInst replaces dummy values within the given sub instruction with their
// real values.
func (fix *fixer) fixSubInst(old *ir.InstSub) *ir.InstSub {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixFSubInst replaces dummy values within the given fsub instruction with their
// real values.
func (fix *fixer) fixFSubInst(old *ir.InstFSub) *ir.InstFSub {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixMulInst replaces dummy values within the given mul instruction with their
// real values.
func (fix *fixer) fixMulInst(old *ir.InstMul) *ir.InstMul {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixFMulInst replaces dummy values within the given fmul instruction with their
// real values.
func (fix *fixer) fixFMulInst(old *ir.InstFMul) *ir.InstFMul {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixUDivInst replaces dummy values within the given udiv instruction with their
// real values.
func (fix *fixer) fixUDivInst(old *ir.InstUDiv) *ir.InstUDiv {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixSDivInst replaces dummy values within the given sdiv instruction with their
// real values.
func (fix *fixer) fixSDivInst(old *ir.InstSDiv) *ir.InstSDiv {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixFDivInst replaces dummy values within the given fdiv instruction with their
// real values.
func (fix *fixer) fixFDivInst(old *ir.InstFDiv) *ir.InstFDiv {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixURemInst replaces dummy values within the given urem instruction with their
// real values.
func (fix *fixer) fixURemInst(old *ir.InstURem) *ir.InstURem {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixSRemInst replaces dummy values within the given srem instruction with their
// real values.
func (fix *fixer) fixSRemInst(old *ir.InstSRem) *ir.InstSRem {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixFRemInst replaces dummy values within the given frem instruction with their
// real values.
func (fix *fixer) fixFRemInst(old *ir.InstFRem) *ir.InstFRem {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// --- [ Bitwise instructions ] ------------------------------------------------

// --- [ Memory instructions ] -------------------------------------------------

// fixLoadInst replaces dummy values within the given load instruction with
// their real values.
func (fix *fixer) fixLoadInst(old *ir.InstLoad) *ir.InstLoad {
	if src, ok := fix.fixValue(old.Src()); ok {
		old.SetSrc(src)
	}
	return old
}

// fixStoreInst replaces dummy values within the given store instruction with
// their real values.
func (fix *fixer) fixStoreInst(old *ir.InstStore) *ir.InstStore {
	if src, ok := fix.fixValue(old.Src()); ok {
		old.SetSrc(src)
	}
	if dst, ok := fix.fixValue(old.Dst()); ok {
		old.SetDst(dst)
	}
	return old
}

// --- [ Conversion instructions ] ---------------------------------------------

// --- [ Other instructions ] --------------------------------------------------

// fixICmpInst replaces dummy values within the given icmp instruction with
// their real values.
func (fix *fixer) fixICmpInst(old *ir.InstICmp) *ir.InstICmp {
	if x, ok := fix.fixValue(old.X()); ok {
		old.SetX(x)
	}
	if y, ok := fix.fixValue(old.Y()); ok {
		old.SetY(y)
	}
	return old
}

// fixCallInst replaces dummy values within the given call instruction with
// their real values.
func (fix *fixer) fixCallInst(old *ir.InstCall) *ir.InstCall {
	var args []value.Value
	for _, arg := range old.Args() {
		arg, _ = fix.fixValue(arg)
		args = append(args, arg)
	}
	old.SetArgs(args)
	return old
}

// fixCallInstDummy replaces the given dummy call instruction with a real call
// instruction, and replaces dummy the instruction with their real values.
func (fix *fixer) fixCallInstDummy(old *instCallDummy) *ir.InstCall {
	callee := fix.getFunc(old.callee)
	var args []value.Value
	for _, arg := range old.args {
		arg, _ = fix.fixValue(arg)
		args = append(args, arg)
	}
	inst := ir.NewCall(callee, args...)
	inst.SetParent(old.parent)
	inst.SetIdent(stripLocal(old.Ident()))
	return inst
}

// === [ Terminators ] =========================================================

// fixTerm replaces dummy values within the given terminator with their real
// values.
func (fix *fixer) fixTerm(term ir.Terminator) ir.Terminator {
	switch term := term.(type) {
	case *ir.TermRet:
		return fix.fixRetTerm(term)
	case *ir.TermBr:
		// nothing to do; contains no values.
		return term
	case *ir.TermCondBr:
		return fix.fixCondBrTerm(term)
	default:
		panic(fmt.Sprintf("support for terminator type %T not yet implemented", term))
	}
}

// fixRetTerm replaces dummy values within the given ret terminator with their
// real values.
func (fix *fixer) fixRetTerm(old *ir.TermRet) *ir.TermRet {
	if x, ok := old.X(); ok {
		if x, ok := fix.fixValue(x); ok {
			old.SetX(x)
		}
	}
	return old
}

// fixBrTermDummy replaces the given dummy unconditional br terminator with a
// real unconditional br terminator, and replaces dummy the terminator with
// their real values.
func (fix *fixer) fixBrTermDummy(old *termBrDummy) *ir.TermBr {
	target := fix.getBloack(old.target)
	term := ir.NewBr(target)
	term.SetParent(old.parent)
	return term
}

// fixCondBrTerm replaces dummy values within the given conditional br
// terminator with their real values.
func (fix *fixer) fixCondBrTerm(old *ir.TermCondBr) *ir.TermCondBr {
	if cond, ok := fix.fixValue(old.Cond()); ok {
		old.SetCond(cond)
	}
	return old
}

// fixCondBrTermDummy replaces the given dummy conditional br terminator with a
// real conditional br terminator, and replaces dummy the terminator with their
// real values.
func (fix *fixer) fixCondBrTermDummy(old *termCondBrDummy) *ir.TermCondBr {
	cond, _ := fix.fixValue(old.cond)
	targetTrue := fix.getBloack(old.targetTrue)
	targetFalse := fix.getBloack(old.targetFalse)
	term := ir.NewCondBr(cond, targetTrue, targetFalse)
	term.SetParent(old.parent)
	return term
}

// ### [ Helper functions ] ####################################################

// stripLocal strips the "%" prefix of the given local identifier.
func stripLocal(s string) string {
	if !strings.HasPrefix(s, "%") {
		panic(fmt.Sprintf(`invalid local identifier %q; missing "%%" prefix`, s))
	}
	return s[1:]
}
