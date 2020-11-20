package pluralforms

// Expression is a plurfalforms expression. Eval evaluates the expression for
// a given n value. Use pluralforms.Compile to generate Expression instances.
type Expression interface {
	Eval(n uint32) int
}

func logic(b bool) int {
	if b {
		return 1
	}
	return 0
}

type notExpr struct {
	sub Expression
}

func (e notExpr) Eval(n uint32) int {
	return logic(e.sub.Eval(n) == 0)
}

type binaryExpr struct {
	left  Expression
	right Expression
}

type orExpr binaryExpr

func (e orExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) != 0 || e.right.Eval(n) != 0)
}

type andExpr binaryExpr

func (e andExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) != 0 && e.right.Eval(n) != 0)
}

type eqExpr binaryExpr

func (e eqExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) == e.right.Eval(n))
}

type neExpr binaryExpr

func (e neExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) != e.right.Eval(n))
}

type ltExpr binaryExpr

func (e ltExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) < e.right.Eval(n))
}

type lteExpr binaryExpr

func (e lteExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) <= e.right.Eval(n))
}

type gtExpr binaryExpr

func (e gtExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) > e.right.Eval(n))
}

type gteExpr binaryExpr

func (e gteExpr) Eval(n uint32) int {
	return logic(e.left.Eval(n) >= e.right.Eval(n))
}

type addExpr binaryExpr

func (e addExpr) Eval(n uint32) int {
	return e.left.Eval(n) + e.right.Eval(n)
}

type subExpr binaryExpr

func (e subExpr) Eval(n uint32) int {
	return e.left.Eval(n) - e.right.Eval(n)
}

type mulExpr binaryExpr

func (e mulExpr) Eval(n uint32) int {
	return e.left.Eval(n) * e.right.Eval(n)
}

type divExpr binaryExpr

func (e divExpr) Eval(n uint32) int {
	return e.left.Eval(n) / e.right.Eval(n)
}

type modExpr binaryExpr

func (e modExpr) Eval(n uint32) int {
	return e.left.Eval(n) % e.right.Eval(n)
}

type ternaryExpr struct {
	test    Expression
	ifTrue  Expression
	ifFalse Expression
}

func (e ternaryExpr) Eval(n uint32) int {
	if e.test.Eval(n) != 0 {
		return e.ifTrue.Eval(n)
	}
	return e.ifFalse.Eval(n)
}

type numberExpr struct {
	value int
}

func (e numberExpr) Eval(n uint32) int {
	return e.value
}

type varExpr struct{}

func (e varExpr) Eval(n uint32) int {
	return int(n)
}
