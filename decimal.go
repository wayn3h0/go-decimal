package decimal

import (
	"bytes"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	errors "github.com/wayn3h0/go-errors"
)

var (
	// max decimal digits allowd for indivisible quotient (exceeding be truncated).
	MaxDecimalDigits = uint(200)
)

// Decimal represents a decimal which can handing fixed precision.
type Decimal struct {
	integer  *big.Int
	exponent int
}

func (d *Decimal) ensureInitialized() {
	if d.integer == nil {
		d.integer = new(big.Int)
	}
}

// Sign returns:
// -1: if d <  0
//  0: if d == 0
// +1: if d >  0
func (d *Decimal) Sign() int {
	d.ensureInitialized()
	return d.integer.Sign()
}

// IsZero reports whether the value of d is equal to zero.
func (d *Decimal) IsZero() bool {
	d.ensureInitialized()
	return d.integer.Sign() == 0
}

// Float32 returns the float32 value nearest to d and a boolean indicating whether is exact.
func (d *Decimal) Float32() (float32, bool) {
	d.ensureInitialized()
	a := new(big.Rat).SetInt(d.integer)
	if d.exponent == 0 {
		return a.Float32()
	}
	b := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), new(big.Int).Abs(big.NewInt(int64(d.exponent))), nil))
	if d.exponent > 0 {
		b.Inv(b)
	}
	z := new(big.Rat).Quo(a, b)
	return z.Float32()
}

// Float64 returns the float64 value nearest to d and a boolean indicating whether is exact.
func (d *Decimal) Float64() (float64, bool) {
	d.ensureInitialized()
	a := new(big.Rat).SetInt(d.integer)
	if d.exponent == 0 {
		return a.Float64()
	}
	b := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), new(big.Int).Abs(big.NewInt(int64(d.exponent))), nil))
	if d.exponent > 0 {
		b.Inv(b)
	}
	z := new(big.Rat).Quo(a, b)
	return z.Float64()
}

// Int64 returns the int64 value nearest to d and a boolean indicating whether is exact.
func (d *Decimal) Int64() (int64, bool) {
	d.ensureInitialized()
	if d.exponent == 0 {
		return d.integer.Int64(), true
	}
	if d.exponent > 0 {
		z := new(big.Int).Mul(d.integer, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.exponent)), nil))
		return z.Int64(), true
	}

	z := new(big.Int).Quo(d.integer, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.exponent*-1)), nil))
	return z.Int64(), false
}

// String converts the floating-point number d to a string.
func (d *Decimal) String() string {
	d.ensureInitialized()
	if d.integer.Cmp(big.NewInt(0)) == 0 { // euqal to zero
		return "0"
	}
	str := d.integer.String()
	if d.exponent == 0 { // value is the integer without exponent
		return str
	}
	if d.exponent > 0 { // value is the integer with exponent
		return str + strings.Repeat("0", d.exponent)
	}

	// has decimal digits
	var buf bytes.Buffer
	if strings.HasPrefix(str, "-") {
		buf.WriteString("-")
		str = str[1:]
	}
	p := len(str) - (d.exponent * -1)
	if p <= 0 {
		buf.WriteString("0.")
		buf.WriteString(strings.Repeat("0", p*-1))
		buf.WriteString(strings.TrimRight(str, "0"))
	} else {
		buf.WriteString(str[:p]) // integer part
		decimals := strings.TrimRight(str[p:], "0")
		if len(decimals) > 0 {
			buf.WriteString(".")
			buf.WriteString(decimals)
		}
	}
	return buf.String()
}

// SetInt64 sets x to y and returns x.
func (x *Decimal) SetInt64(y int64) *Decimal {
	x.ensureInitialized()
	x.integer.SetInt64(y)
	x.exponent = 0
	return x
}

var (
	_DecimalPattern = regexp.MustCompile(`^([-+]?\d+)(\.(\d+))?([eE]([-+]?\d+))?$`)
)

// SetString sets x to the value of y and returns x and a boolean indicating success.
// If the operation failed, the value of d is undefined but the returned value is nil.
func (x *Decimal) SetString(y string) (*Decimal, bool) {
	x.ensureInitialized()
	matches := _DecimalPattern.FindStringSubmatch(y)
	if len(matches) != 6 {
		return nil, false
	}
	decimals := strings.TrimRight(matches[3], "0")
	integer := matches[1] + decimals
	exponent := len(decimals) * -1
	if len(matches[5]) > 0 {
		exp, _ := strconv.ParseInt(matches[5], 10, 64)
		exponent += int(exp)
	}
	x.integer.SetString(integer, 10)
	x.exponent = exponent

	return x, true
}

// SetFloat64 sets x to y and returns x.
func (x *Decimal) SetFloat64(y float64) *Decimal {
	x.ensureInitialized()
	x.SetString(strconv.FormatFloat(y, 'f', -1, 64))
	return x
}

// Copy sets x to y and returns x. y is not changed.
func (x *Decimal) Copy(y *Decimal) *Decimal {
	x.ensureInitialized()
	y.ensureInitialized()
	x.integer.Set(y.integer)
	x.exponent = y.exponent
	return x
}

// Abs sets d to the value |d| (the absolute value of d) and returns d.
func (d *Decimal) Abs() *Decimal {
	d.ensureInitialized()
	d.integer.Abs(d.integer)
	return d
}

// Neg sets d to the value of d with its sign negated, and returns d.
func (d *Decimal) Neg() *Decimal {
	d.ensureInitialized()
	d.integer.Neg(d.integer)
	return d
}

func (x *Decimal) align(y *Decimal) {
	if x.exponent != y.exponent {
		diff := new(big.Int).Abs(new(big.Int).Sub(new(big.Int).Abs(big.NewInt(int64(x.exponent))), new(big.Int).Abs(big.NewInt(int64(y.exponent)))))
		if x.exponent > y.exponent {
			x.integer.Mul(x.integer, new(big.Int).Exp(big.NewInt(10), diff, nil))
			x.exponent = y.exponent
		} else {
			y.integer.Mul(y.integer, new(big.Int).Exp(big.NewInt(10), diff, nil))
			y.exponent = x.exponent
		}
	}
}

// Cmp compares x and y and returns:
// -1 if d < y
//  0 if d == y (includes: -0 == 0, -Inf == -Inf, and +Inf == +Inf)
// +1 if d > y
func (x *Decimal) Cmp(y *Decimal) int {
	x.ensureInitialized()
	y.ensureInitialized()
	x.align(y)
	return x.integer.Cmp(y.integer)
}

// Add sets d to the sum of d and y and returns x.
func (x *Decimal) Add(y *Decimal) *Decimal {
	x.ensureInitialized()
	y.ensureInitialized()
	x.align(y)
	x.integer.Add(x.integer, y.integer)
	return x
}

// Sub sets d to the difference x-y and returns x.
func (x *Decimal) Sub(y *Decimal) *Decimal {
	x.ensureInitialized()
	y.ensureInitialized()
	x.align(y)
	x.integer.Sub(x.integer, y.integer)
	return x
}

// Mul sets x to the product x*y and returns x.
func (x *Decimal) Mul(y *Decimal) *Decimal {
	x.ensureInitialized()
	y.ensureInitialized()
	if y.integer.Sign() == 0 { // *0
		x.integer.SetInt64(0)
		x.exponent = 0
		return x
	}
	x.integer.Mul(x.integer, y.integer)
	x.exponent += y.exponent
	return x
}

// Quo sets x to the quotient x/y and return x.
// Please set MaxDecimalDigitis for indivisible quotient.
func (x *Decimal) Quo(y *Decimal) *Decimal {
	x.ensureInitialized()
	y.ensureInitialized()
	if y.integer.Sign() == 0 { // /0
		x.integer.SetInt64(0)
		x.exponent = 0
		return x
	}
	// modulus x%y == 0
	if z, r := new(big.Int).QuoRem(x.integer, y.integer, new(big.Int)); r.Sign() == 0 {
		x.integer = z
		x.exponent -= y.exponent
		return x
	}
	// modulus x%y > 0
	var buf bytes.Buffer
	if x.integer.Sign()*y.integer.Sign() == -1 {
		buf.WriteString("-")
	}
	xi := new(big.Int).Abs(x.integer)
	yi := new(big.Int).Abs(y.integer)
	exp := x.exponent - y.exponent
	z, r := new(big.Int).QuoRem(xi, yi, new(big.Int))
	buf.WriteString(z.String())
	for r.Sign() != 0 && exp*-1 < int(MaxDecimalDigits) {
		r.Mul(r, big.NewInt(10))
		z, r = new(big.Int).QuoRem(r, yi, new(big.Int))
		buf.WriteString(z.String())
		exp -= 1
	}
	str := fmt.Sprintf("%se%d", buf.String(), exp)
	x.SetString(str)
	return x
}

// Div is same to Quo.
func (x *Decimal) Div(y *Decimal) *Decimal {
	return x.Quo(y)
}

// RoundToNearestEven rounds (IEEE 754-2008, round to nearest, ties to even) the floating-point number x with given precision (the number of digits after the decimal point).
func (d *Decimal) RoundToNearestEven(precision uint) *Decimal {
	d.ensureInitialized()
	prec := int(precision)
	if d.IsZero() || d.exponent > 0 || d.exponent*-1 <= prec { // rounding needless
		return d
	}

	str := d.integer.String()
	var sign, part1, part2 string
	if strings.HasPrefix(str, "-") {
		sign = "-"
		str = str[1:]
	}
	if len(str) < d.exponent*-1 {
		str = strings.Repeat("0", (d.exponent*-1)-len(str)+1) + str
	}
	part1 = str[:len(str)+d.exponent]
	part2 = str[len(part1):]
	isRoundUp := false
	switch part2[prec : prec+1] {
	case "6", "7", "8", "9":
		isRoundUp = true
	case "5":
		if len(part2) > prec+1 { // found decimals back of "5"
			isRoundUp = true
		} else {
			var neighbor string
			if prec == 0 { // get neighbor from integer part
				neighbor = part1[len(part1)-1:]
			} else {
				neighbor = part2[prec-1 : prec]
			}
			switch neighbor {
			case "1", "3", "5", "7", "9":
				isRoundUp = true
			}
		}
	}
	z, _ := new(big.Int).SetString(sign+part1+part2[:prec], 10)
	if isRoundUp {
		z.Add(z, big.NewInt(int64(d.integer.Sign())))
	}
	d.integer = z
	d.exponent = prec * -1
	return d
}

// Round is short to RoundToNearestEven.
func (d *Decimal) Round(precision uint) *Decimal {
	return d.RoundToNearestEven(precision)
}

// RoundToNearestAway rounds (IEEE 754-2008, round to nearest, ties away from zero) the floating-point number x with given precision (the number of digits after the decimal point).
func (d *Decimal) RoundToNearestAway(precision uint) *Decimal {
	d.ensureInitialized()
	prec := int(precision)
	if d.IsZero() || d.exponent > 0 || d.exponent*-1 <= prec { // rounding needless
		return d
	}

	diff := new(big.Int).Sub(new(big.Int).Abs(big.NewInt(int64(d.exponent))), big.NewInt(int64(prec+1)))
	d.integer.Quo(d.integer, new(big.Int).Exp(big.NewInt(10), diff, nil))
	factor := big.NewInt(int64(5))
	if d.integer.Sign() == -1 {
		factor.Neg(factor)
	}
	d.integer.Add(d.integer, factor)
	d.integer.Quo(d.integer, big.NewInt(10))
	d.exponent = prec * -1
	return d
}

// RoundToZero rounds (IEEE 754-2008, round towards zero) the floating-point number x with given precision (the number of digits after the decimal point).
func (d *Decimal) RoundToZero(precision uint) *Decimal {
	d.ensureInitialized()
	prec := int(precision)
	if d.IsZero() || d.exponent > 0 || d.exponent*-1 <= prec { // rounding needless
		return d
	}

	diff := new(big.Int).Sub(new(big.Int).Abs(big.NewInt(int64(d.exponent))), big.NewInt(int64(prec)))
	d.integer.Quo(d.integer, new(big.Int).Exp(big.NewInt(10), diff, nil))
	d.exponent = prec * -1
	return d
}

// Truncate is same as RoundToZero.
func (d *Decimal) Truncate(precision uint) *Decimal {
	return d.RoundToZero(precision)
}

// RoundDown is same as RoundToZero.
func (d *Decimal) RoundDown(precision uint) *Decimal {
	return d.RoundToZero(precision)
}

// RoundAwayFromZero rounds (no IEEE 754-2008, round away from zero) to floating-point number d with given precision.
func (d *Decimal) RoundAwayFromZero(precision uint) *Decimal {
	d.ensureInitialized()
	prec := int(precision)
	if d.IsZero() || d.exponent > 0 || d.exponent*-1 <= prec { // rounding needless
		return d
	}

	sign := d.integer.Sign()
	diff := new(big.Int).Sub(new(big.Int).Abs(big.NewInt(int64(d.exponent))), big.NewInt(int64(prec)))
	if _, r := d.integer.QuoRem(d.integer, new(big.Int).Exp(big.NewInt(10), diff, nil), new(big.Int)); r.Sign() != 0 {
		d.integer.Add(d.integer, big.NewInt(int64(1*sign))) // round up
	}
	d.exponent = prec * -1
	return d
}

// RoundUp is same as RoundAwayFromZero.
func (d *Decimal) RoundUp(precision uint) *Decimal {
	return d.RoundAwayFromZero(precision)
}

// New returns a new decimal.
func New(number float64) *Decimal {
	return new(Decimal).SetFloat64(number)
}

// Parse returns a new decimal by parsing decimal string.
func Parse(str string) (*Decimal, error) {
	if d, ok := new(Decimal).SetString(str); ok {
		return d, nil
	}
	return nil, errors.Newf("decimal string %q is invalid", str)
}

// MustParse is similar to ParseDecimal but panics if error occurred.
func MustParse(str string) *Decimal {
	d, err := Parse(str)
	if err != nil {
		panic(err)
	}
	return d
}
