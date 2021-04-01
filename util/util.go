package util

import (
	"fmt"
	"math"

	"github.com/uzudil/bscript/bscript"
)

func NewFunctionCall(functionName string, values ...*bscript.Value) *bscript.Variable {
	args := make([]*bscript.Expression, len(values))
	for i, v := range values {
		args[i] = &bscript.Expression{
			BoolTerm: &bscript.BoolTerm{
				Left: &bscript.Cmp{
					Left: &bscript.Term{
						Left: &bscript.Factor{
							Base: v,
						},
					},
				},
			},
		}
	}

	return &bscript.Variable{
		Variable: functionName,
		Suffixes: []*bscript.VariableSuffix{
			{
				CallParams: &bscript.CallParams{
					Args: args,
				},
			},
		},
	}
}

var trueValue string = "true"
var falseValue string = "false"

func toValue(v interface{}) *bscript.Value {
	value := &bscript.Value{}
	if f, ok := v.(float64); ok {
		value.Number = &bscript.SignedNumber{}
		value.Number.Number = f
	} else {
		if s, ok := v.(string); ok {
			value.String = &s
		} else {
			if b, ok := v.(bool); ok {
				if b {
					value.Boolean = &trueValue
				} else {
					value.Boolean = &falseValue
				}
			} else {
				if a, ok := v.(*([]interface{})); ok {
					value.Array = &bscript.Array{}
					for _, e := range *a {
						expr := toExpression(toValue(e))
						if value.Array.LeftValue == nil {
							value.Array.LeftValue = expr
						} else {
							value.Array.RightValues = append(value.Array.RightValues, expr)
						}
					}
				} else {
					if m, ok := v.(map[string]interface{}); ok {
						value.Map = ToBscriptMap(m)
					} else {
						panic(fmt.Sprintf("Don't know how to convert value type: %v", v))
					}
				}
			}
		}
	}
	return value
}

func toExpression(value *bscript.Value) *bscript.Expression {
	return &bscript.Expression{
		BoolTerm: &bscript.BoolTerm{
			Left: &bscript.Cmp{
				Left: &bscript.Term{
					Left: &bscript.Factor{
						Base: value,
					},
				},
			},
		},
	}
}

func ToBscriptMap(d map[string]interface{}) *bscript.Map {
	m := &bscript.Map{}
	for k, v := range d {
		expr := toExpression(toValue(v))
		nvp := &bscript.NameValuePair{
			Name:  k,
			Value: expr,
		}
		if m.LeftNameValuePair == nil {
			m.LeftNameValuePair = nvp
		} else {
			m.RightNameValuePairs = append(m.RightNameValuePairs, nvp)
		}
	}
	return m
}

func Linear(a, b, percent float32) float32 {
	return a + (b-a)*percent
}

func Clamp(value, min, max float32) float32 {
	return float32(math.Max(float64(min), math.Min(float64(value), float64(max))))
}
