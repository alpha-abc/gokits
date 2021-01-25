package validatorv1

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

/**
 * tag默认规则: 以逗号分割, e.g. valid:{rule1},{rule2},{rule3<regex1>}...
 * - 普通规则: 非逗号字符
 * - 正则规则: 以符号^开始, 以符号$结束
 */

// tag相关定义
const (
	TagValid   = "valid"
	TagRegex   = "regex"
	TagMessage = "message"

	EmptyRule    = ""
	EmptyRegex   = ""
	EmptyMessage = ""

	TagSeparatorChar = ','
	TagSeparator     = ","
)

const (
	// RuleTypeNormal 普通类型规则
	RuleTypeNormal = 1
	// RuleTypeRegex 正则类型规则
	RuleTypeRegex = 2
)

// 规则名称
const (
	regex    = "regex"
	required = "required"
	next     = "->"
)

// UnsupportedTypeError 未支持的类型
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "validator: unsupported type: " + e.Type.String()
}

// ValidationError ...
type ValidationError struct {
	Key  string
	Text string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Text)
}

type element struct {
	parentKind reflect.Kind

	key   string      // 变量信息
	value interface{} // 值

	rules   []rule // 规则
	message string // 变量错误提示信息
}

type rule struct {
	typ uint
	val string
}

func RuleExtractFromTag(tagValue string) []rule {
	var text = []rune(tagValue)
	var textLen = len(text)

	var separatorArr = make([]bool, textLen)
	var cacheIdxArr = make([]int, textLen)
	var idx = 0

	// 是否处于匹配状态, ^和$匹配
	var b = false

	for i, c := range text {
		if c == TagSeparatorChar {
			if !b {
				separatorArr[i] = true
			} else {
				cacheIdxArr[idx] = i
				idx++
			}
			continue
		}

		if c == '^' {
			if b {
				for j := 0; j < idx; j++ {
					separatorArr[cacheIdxArr[j]] = true
				}

				idx = 0
			} else {
				b = true
			}
			continue
		}

		if c == '$' {
			if b {
				idx = 0
				b = false
			}
		}
	}

	var rules []rule
	var left = -1
	for i, b := range separatorArr {
		if b {
			var r = strings.Trim(string(text[left+1:i]), " ")

			if len(r) >= 2 && r[0] == '^' && r[len(r)-1] == '$' {
				rules = append(rules, rule{typ: RuleTypeRegex, val: r})
			} else if len(r) > 0 {
				rules = append(rules, rule{typ: RuleTypeNormal, val: r})
			}

			left = i
		}

		if i == textLen-1 && left != i {
			var r = strings.Trim(string(text[left+1:textLen]), " ")

			if len(r) >= 2 && r[0] == '^' && r[len(r)-1] == '$' {
				rules = append(rules, rule{typ: RuleTypeRegex, val: r})
			} else if len(r) > 0 {
				rules = append(rules, rule{typ: RuleTypeNormal, val: r})
			}
		}
	}

	return rules
}

//type ruleFunc func(key string, v reflect.Value, ruleText string, message string) error
type ruleFunc func(v reflect.Value, ruleText string) error

func ruleRequired(v reflect.Value, ruleText string) error {
	if v.IsNil() {
		return fmt.Errorf("%s is required", v.String())
	}
	return nil
}

func ruleRegex(val reflect.Value, ruleText string) error {
	var regex = ruleText

	var r, err = regexp.Compile(regex)
	if err != nil {
		return err
	}

	var valStr = fmt.Sprintf("%v", val.Interface())

	if !r.Match([]byte(valStr)) {
		return fmt.Errorf(`regex: "%s" not match the value: "%s"`, regex, valStr)
	}

	return nil
}

var kindRulesMap = map[reflect.Kind]map[string]ruleFunc{
	reflect.Ptr: {
		required: ruleRequired,
	},
	reflect.Array: {
		required: ruleRequired,
	},
	reflect.Slice: {
		required: ruleRequired,
	},
	reflect.Bool:    {regex: ruleRegex},
	reflect.Int:     {regex: ruleRegex},
	reflect.Int8:    {regex: ruleRegex},
	reflect.Int16:   {regex: ruleRegex},
	reflect.Int32:   {regex: ruleRegex},
	reflect.Int64:   {regex: ruleRegex},
	reflect.Uint:    {regex: ruleRegex},
	reflect.Uint8:   {regex: ruleRegex},
	reflect.Uint16:  {regex: ruleRegex},
	reflect.Uint32:  {regex: ruleRegex},
	reflect.Uint64:  {regex: ruleRegex},
	reflect.Float32: {regex: ruleRegex},
	reflect.Float64: {regex: ruleRegex},
	reflect.String:  {regex: ruleRegex},
}

func validate(v reflect.Value, rules []rule) error {
	var rs, ok = kindRulesMap[v.Kind()]
	if !ok {
		return nil
	}

	for _, rule := range rules {
		var r = rule.val
		if rule.typ == RuleTypeRegex {
			r = regex
		}

		var ruleFn, ok = rs[r]
		if !ok {
			continue
		}

		if err := ruleFn(v, rule.val); err != nil {
			return err
		}
	}

	return nil
}

// Validate 验证
func Validate(v interface{}) []error {
	var errs []error

	var lst = list.New()
	lst.PushBack(&element{
		parentKind: reflect.Invalid,

		value:   v,
		rules:   []rule{{typ: RuleTypeNormal, val: required}},
		key:     "",
		message: "",
	})

	for lst.Len() > 0 {
		var first = lst.Front()
		lst.Remove(first)

		var ele = first.Value.(*element)
		var value = reflect.ValueOf(ele.value)

		if err := validate(value, ele.rules); err != nil {
			var e = new(ValidationError)
			e.Key = ele.key

			if ele.message == EmptyMessage {
				e.Text = err.Error()
			} else {
				e.Text = ele.message
			}

			errs = append(errs, e)
		}

		switch value.Kind() {
		case reflect.Struct:

			for i := 0; i < value.NumField(); i++ {
				var fieldValue = value.Field(i)
				var fieldType = value.Type().Field(i)

				var tagValid, _ = fieldType.Tag.Lookup(TagValid)
				var tagMessage, _ = fieldType.Tag.Lookup(TagMessage)

				var fmtStr = "%s.%s.%s"
				if ele.parentKind == reflect.Ptr {
					fmtStr = "%s%s.%s"
				}

				lst.PushBack(&element{
					parentKind: value.Kind(),
					value:      fieldValue.Interface(),
					rules:      RuleExtractFromTag(tagValid),
					key:        fmt.Sprintf(fmtStr, ele.key, value.Type().Name(), fieldType.Name),
					message:    tagMessage,
				})
			}

		case reflect.Array, reflect.Slice:
			// 数组, 切片需要 `next` 规则
			for i := 0; i < value.Len(); i++ {
				lst.PushBack(&element{
					parentKind: value.Kind(),
					value:      value.Index(i).Interface(),
					rules:      ele.rules,
					key:        fmt.Sprintf("%s[%d]", ele.key, i),
					message:    ele.message,
				})
			}

		case reflect.Ptr:
			if !value.IsNil() {
				lst.PushBack(&element{
					parentKind: value.Kind(),
					value:      value.Elem().Interface(),
					rules:      ele.rules,
					key:        fmt.Sprintf("%s*", ele.key),
					message:    ele.message,
				})
			}

		case reflect.Bool,
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Float32,
			reflect.Float64,
			reflect.String:

			// 验证结束, 不再有新元素添加

			// 后续扩充处理以下类型
		case reflect.Map:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.Complex64, reflect.Complex128:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.Uintptr:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.Chan:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.Func:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.Interface:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.UnsafePointer:
			errs = append(errs, &UnsupportedTypeError{value.Type()})
		case reflect.Invalid:
			errs = append(errs, errors.New("invalid type error"))
		default:
			errs = append(errs, errors.New("unsupport error"))
		}

	}
	return errs
}

// min=10,range=1:2:3,/,email
// 用到的特殊符号 , : /

// Validator ..
type Validator struct {
	mp map[reflect.Kind]map[string]Ruler
}

// Ruler 规则
type Ruler interface {
	RuleDescer
	RuleFuncer
}

// RuleDescer ..
type RuleDescer interface {
	Desc() string
}

// RuleFuncer ..
type RuleFuncer interface {
	Verify(reflect.Value, []string) error
}

// TagRule 从tag标签中解析生成
type TagRule struct {
	Name   string
	Params []string
}

// SetValidationFunc 设置验证函数
func (v *Validator) SetValidationFunc(tp reflect.Kind, ruleName string, rf Ruler) error {
	if tp == reflect.Invalid {
		return errors.New("invalid reflect kind")
	}

	if ruleName == "" {
		return errors.New("rule name cannot be empty")
	}

	if rf == nil {
		if _, ok := v.mp[tp]; ok {
			delete(v.mp[tp], ruleName)
		}

		return nil
	}

	if _, ok := v.mp[tp]; !ok {
		v.mp[tp] = make(map[string]Ruler)
	}

	v.mp[tp][ruleName] = rf
	return nil
}

// Verify 参数验证
func (v *Validator) Verify(rv reflect.Value, ruleName string, params []string) error {
	var rule, ok = v.mp[rv.Kind()]
	if !ok {
		// 函数不支持的类型校验
		return nil
	}

	var r, ok1 = rule[ruleName]
	if !ok1 {
		// 未知/未实现的验证函数
		return nil
	}

	return r.Verify(rv, params)
}

var commasPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*),`)
var semicolonPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*);`)

var equalPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*)=`)
var slashPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*)/`)

// 特殊字符
const (
	SymbolSemicolon        string = ";"
	SymbolSemicolonEscaped string = "\\;"
	SymbolCommas           string = ","
	SymbolCommasEscaped    string = "\\,"
	SymbolEqual            string = "="
	SymbolSpace            string = " "
)

// ExtractFromTag 一共有3个特殊符号, `;`, `,` 和 `=`, 其中分号和逗号如果当成普通字符出现, 需要转义(\\; \\,)
func (v *Validator) ExtractFromTag(tagStr string) (tagRule []TagRule) {
	var semIdxArr = semicolonPattern.FindAllStringIndex(tagStr, -1)

	// 处理分隔符 `;`
	var rawRules []string
	var lastSemIdx = 0
	for _, arr := range semIdxArr {
		rawRules = append(rawRules, tagStr[lastSemIdx:arr[1]-1])
		lastSemIdx = arr[1]
	}
	rawRules = append(rawRules, tagStr[lastSemIdx:])

	for _, rawRule := range rawRules {
		if strings.Trim(rawRule, SymbolSpace) == "" {
			continue
		}

		var tr = TagRule{}

		var rule = strings.ReplaceAll(rawRule, SymbolSemicolonEscaped, SymbolSemicolon)

		var nameParamsArr = strings.SplitN(rule, SymbolEqual, 2)
		var name = strings.Trim(nameParamsArr[0], SymbolSpace)

		// 以等号出现的情况
		if name == "" {
			continue
		}

		tr.Name = name

		if len(nameParamsArr) > 1 {
			var rawParams = nameParamsArr[1]
			var commasIdxArr = commasPattern.FindAllStringIndex(rawParams, -1)

			var params []string
			var lastCommasIdx = 0
			for _, arr := range commasIdxArr {
				params = append(params, rawParams[lastCommasIdx:arr[1]-1])
				lastCommasIdx = arr[1]
			}
			params = append(params, rawParams[lastCommasIdx:])

			for i := 0; i < len(params); i++ {
				var p = strings.ReplaceAll(params[i], SymbolCommasEscaped, SymbolCommas)
				params[i] = p
			}

			tr.Params = params
		}

		tagRule = append(tagRule, tr)
	}

	return
}

// Validate ..
func (v *Validator) Validate(vf interface{}) {

}

type ErrorArr []error
type ValueMap map[string]ErrorArr

type Value struct {
	pKind reflect.Kind
}

func (v *Validator) Recursively(rv reflect.Value, idx int) {

	switch rv.Kind() {
	case reflect.Interface, reflect.Ptr:
		if rv.IsNil() {
			return
		}

		v.Recursively(rv.Elem())

	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {

		}
	}

}
