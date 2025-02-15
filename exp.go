package decimal128

import "math/bits"

// Exp returns e**d, the base-e exponential of d.
func Exp(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	if d.IsZero() {
		return one(false)
	}

	dSig, dExp := d.decompose()
	dExp -= exponentBias
	l10 := dSig.log10()

	if int(dExp) > 5-l10 {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	res, trunc := decomposed128{
		sig: dSig,
		exp: dExp,
	}.epow(int16(l10), int8(0))

	if res.exp > maxUnbiasedExponent+39 {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	if d.Signbit() {
		res, trunc = decomposed128{
			sig: uint128{1, 0},
			exp: 0,
		}.quo(res, trunc)
	}

	sig, exp := DefaultRoundingMode.reduce128(false, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	return compose(false, sig, exp)
}

// Exp10 returns 10**d, the base-10 exponential of d.
func Exp10(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	if d.IsZero() {
		return one(false)
	}

	dSig, dExp := d.decompose()
	dExp -= exponentBias
	l10 := dSig.log10()

	if int(dExp) > 4-l10 {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	var dSigInt uint
	if l10+int(dExp) >= 0 {
		sig := dSig
		exp := dExp
		dSig = uint128{}

		for exp < 0 {
			var rem uint64
			sig, rem = sig.div10()

			dSig = dSig.mul64(10)
			dSig = dSig.add64(rem)
			exp++
		}

		dSigInt = uint(sig[0])

		for exp > 0 {
			dSigInt *= 10
			exp--
		}

		if dSigInt > maxUnbiasedExponent+39 {
			if d.Signbit() {
				return zero(false)
			}

			return inf(false)
		}

		sig = dSig
		dSig = uint128{}
		dExp = 0

		for sig != (uint128{}) {
			var rem uint64
			sig, rem = sig.div10()

			dSig = dSig.mul64(10)
			dSig = dSig.add64(rem)
			dExp--
		}
	}

	var res decomposed128
	var trunc int8

	var sigInt uint128
	var expInt int16

	if dSigInt != 0 {
		sigInt = uint128{1, 0}

		for dSigInt > maxUnbiasedExponent {
			sigInt = sigInt.mul64(10)
			dSigInt--
		}

		expInt = int16(dSigInt)
	}

	if dSig != (uint128{}) {
		res, trunc = decomposed128{
			sig: dSig,
			exp: dExp,
		}.mul(ln10, int8(0))

		res, trunc = res.epow(int16(l10), trunc)

		if res.exp > maxUnbiasedExponent+39 {
			if d.Signbit() {
				return zero(false)
			}

			return inf(false)
		}

		if expInt != 0 {
			res.exp += expInt
		}
	} else {
		res = decomposed128{
			sig: uint128{1, 0},
			exp: expInt,
		}
	}

	if res.exp > maxUnbiasedExponent+39 {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	if d.Signbit() {
		res, trunc = decomposed128{
			sig: uint128{1, 0},
			exp: 0,
		}.quo(res, trunc)
	}

	sig, exp := DefaultRoundingMode.reduce128(false, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	return compose(false, sig, exp)
}

// Exp2 returns 2**d, the base-2 exponential of d.
func Exp2(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	if d.IsZero() {
		return one(false)
	}

	dSig, dExp := d.decompose()
	dExp -= exponentBias
	l10 := dSig.log10()

	if int(dExp) > 5-l10 {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	var dSigInt uint
	if l10+int(dExp) >= 0 {
		sig := dSig
		exp := dExp
		dSig = uint128{}

		for exp < 0 {
			var rem uint64
			sig, rem = sig.div10()

			dSig = dSig.mul64(10)
			dSig = dSig.add64(rem)
			exp++
		}

		dSigInt = uint(sig[0])

		for exp > 0 {
			dSigInt *= 10
			exp--
		}

		sig = dSig
		dSig = uint128{}
		dExp = 0

		for sig != (uint128{}) {
			var rem uint64
			sig, rem = sig.div10()

			dSig = dSig.mul64(10)
			dSig = dSig.add64(rem)
			dExp--
		}
	}

	var res decomposed128
	var trunc int8

	var sigInt uint128
	var expInt int16

	if dSigInt != 0 {
		if dSigInt > exponentBias+maxDigits {
			if d.Signbit() {
				return zero(false)
			}

			return inf(false)
		}

		shift := dSigInt

		if shift < 64 {
			sigInt[0] = 1 << shift
		} else if shift < 128 {
			sigInt[1] = 1 << (shift - 64)
		} else {
			var sigInt256 uint256
			if shift < 192 {
				sigInt256[2] = 1 << (shift - 128)
			} else if shift < 256 {
				sigInt256[3] = 1 << (shift - 192)
			} else {
				sigInt256[3] = 0x8000_0000_0000_0000
				shift -= 255

				for shift > 0 {
					var rem uint64
					sigInt256, rem = sigInt256.div10()
					expInt++

					if rem != 0 {
						trunc = 1
					}

					zeros := uint(bits.LeadingZeros64(sigInt256[3]))
					if shift > zeros {
						sigInt256 = sigInt256.lsh(zeros)
						shift -= zeros
					} else {
						sigInt256 = sigInt256.lsh(shift)
						break
					}
				}
			}

			for sigInt256[3] > 0 {
				var rem uint64
				sigInt256, rem = sigInt256.div1e19()
				expInt += 19

				if rem != 0 {
					trunc = 1
				}
			}

			sigInt192 := uint192{sigInt256[0], sigInt256[1], sigInt256[2]}

			for sigInt192[2] >= 0x0000_0000_0000_ffff {
				var rem uint64
				sigInt192, rem = sigInt192.div10000()
				expInt += 4

				if rem != 0 {
					trunc = 1
				}
			}

			for sigInt192[2] > 0 {
				var rem uint64
				sigInt192, rem = sigInt192.div10()
				expInt++

				if rem != 0 {
					trunc = 1
				}
			}

			sigInt = uint128{sigInt192[0], sigInt192[1]}
		}
	}

	if dSig != (uint128{}) {
		res, trunc = decomposed128{
			sig: dSig,
			exp: dExp,
		}.mul(ln2, int8(0))

		res, trunc = res.epow(int16(l10), trunc)

		if res.exp > maxUnbiasedExponent+39 {
			if d.Signbit() {
				return zero(false)
			}

			return inf(false)
		}

		if dSigInt != 0 {
			res, trunc = decomposed128{
				sig: sigInt,
				exp: expInt,
			}.mul(res, trunc)
		}
	} else {
		res = decomposed128{
			sig: sigInt,
			exp: expInt,
		}
	}

	if res.exp > maxUnbiasedExponent+maxDigits {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	if d.Signbit() {
		res, trunc = decomposed128{
			sig: uint128{1, 0},
			exp: 0,
		}.quo(res, trunc)
	}

	sig, exp := DefaultRoundingMode.reduce128(false, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		if d.Signbit() {
			return zero(false)
		}

		return inf(false)
	}

	return compose(false, sig, exp)
}

// Log returns the natural logarithm of d.
func Log(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return nan(payloadOpLog, payloadValNegInfinite, 0)
		}

		return inf(false)
	}

	if d.IsZero() {
		return inf(true)
	}

	if d.Signbit() {
		return nan(payloadOpLog, payloadValNegFinite, 0)
	}

	dSig, dExp := d.decompose()

	neg, res, trunc := decomposed128{
		sig: dSig,
		exp: dExp - exponentBias,
	}.log()

	sig, exp := DefaultRoundingMode.reduce128(neg, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		return inf(neg)
	}

	return compose(neg, sig, exp)
}

// Log10 returns the decimal logarithm of d.
func Log10(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return nan(payloadOpLog10, payloadValNegInfinite, 0)
		}

		return inf(false)
	}

	if d.IsZero() {
		return inf(true)
	}

	if d.Signbit() {
		return nan(payloadOpLog10, payloadValNegFinite, 0)
	}

	dSig, dExp := d.decompose()

	neg, res, trunc := decomposed128{
		sig: dSig,
		exp: dExp - exponentBias,
	}.log()

	res, trunc = res.mul(invLn10, trunc)

	sig, exp := DefaultRoundingMode.reduce128(neg, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		return inf(neg)
	}

	return compose(neg, sig, exp)
}

// Log2 returns the binary logarithm of d.
func Log2(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return nan(payloadOpLog2, payloadValNegInfinite, 0)
		}

		return inf(false)
	}

	if d.IsZero() {
		return inf(true)
	}

	if d.Signbit() {
		return nan(payloadOpLog2, payloadValNegFinite, 0)
	}

	dSig, dExp := d.decompose()

	neg, res, trunc := decomposed128{
		sig: dSig,
		exp: dExp - exponentBias,
	}.log()

	res, trunc = res.mul(invLn2, trunc)

	sig, exp := DefaultRoundingMode.reduce128(neg, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		return inf(neg)
	}

	return compose(neg, sig, exp)
}

// Sqrt returns the square root of d.
func Sqrt(d Decimal) Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return d
		}

		if d.Signbit() {
			return nan(payloadOpSqrt, payloadValNegInfinite, 0)
		}

		return d
	}

	if d.IsZero() {
		return d
	}

	if d.Signbit() {
		return nan(payloadOpSqrt, payloadValNegFinite, 0)
	}

	dSig, dExp := d.decompose()
	l10 := int16(dSig.log10())
	dExp = (dExp - exponentBias) + l10

	var add decomposed128
	var mul decomposed128
	var nrm decomposed128
	if dExp&1 == 0 {
		add = decomposed128{
			sig: uint128{259, 0},
			exp: -3,
		}

		mul = decomposed128{
			sig: uint128{819, 0},
			exp: -3,
		}

		nrm = decomposed128{
			sig: dSig,
			exp: -l10,
		}
	} else {
		add = decomposed128{
			sig: uint128{819, 0},
			exp: -4,
		}

		mul = decomposed128{
			sig: uint128{259, 0},
			exp: -2,
		}

		nrm = decomposed128{
			sig: dSig,
			exp: -l10 - 1,
		}

		dExp++
	}

	res, trunc := nrm.mul(mul, int8(0))
	res, trunc = res.add(add, trunc)

	var tmp decomposed128
	half := decomposed128{
		sig: uint128{5, 0},
		exp: -1,
	}

	for i := 0; i < 8; i++ {
		tmp, trunc = nrm.quo(res, trunc)
		res, trunc = res.add(tmp, trunc)
		res, trunc = half.mul(res, trunc)
	}

	res.exp += dExp / 2
	sig, exp := DefaultRoundingMode.reduce128(false, res.sig, res.exp+exponentBias, trunc)

	if exp > maxBiasedExponent {
		return inf(false)
	}

	return compose(false, sig, exp)
}
