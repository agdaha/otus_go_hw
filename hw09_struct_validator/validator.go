package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotStruct        = errors.New("value must be a struct")
	ErrUnsupportedType  = errors.New("unsupported field type")
	ErrInvalidValidator = errors.New("invalid validator")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	msgs := make([]string, 0, len(v))
	for _, e := range v {
		msgs = append(msgs, fmt.Sprintf("field %q: %v", e.Field, e.Err))
	}
	return strings.Join(msgs, "; ")
}

func Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	return validateStruct(rv, "")
}

func validateStruct(rv reflect.Value, prefix string) error {
	rt := rv.Type()

	var errs ValidationErrors

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		if !field.IsExported() {
			continue
		}

		tag, ok := field.Tag.Lookup("validate")
		if !ok || tag == "" {
			continue
		}

		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		if tag == "nested" {
			kind := fv.Kind()
			if kind == reflect.Ptr {
				fv = fv.Elem()
				kind = fv.Kind()
			}
			if kind != reflect.Struct {
				return fmt.Errorf("%w: field %q has `nested` tag but is not a struct", ErrInvalidValidator, fieldName)
			}

			nestedErr := validateStruct(fv, fieldName)
			if nestedErr == nil {
				continue
			}

			var ve ValidationErrors
			if errors.As(nestedErr, &ve) {
				errs = append(errs, ve...)
				continue
			}

			return nestedErr
		}
		rules := strings.Split(tag, "|")

		fieldErrs, err := applyValidateRules(fv, fieldName, rules)
		if err != nil {
			return err
		}
		errs = append(errs, fieldErrs...)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func applyValidateRules(fv reflect.Value, fieldName string, rules []string) (ValidationErrors, error) {
	for fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			return nil, nil
		}
		fv = fv.Elem()
	}

	kind := fv.Kind()

	switch kind {
	case reflect.String:
		return applyStringRules(fv.String(), fieldName, rules)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return applyIntRules(fv.Int(), fieldName, rules)

	case reflect.Slice:
		return applySliceRules(fv, fieldName, rules)

	case reflect.Invalid,
		reflect.Bool,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Struct,
		reflect.Pointer,
		reflect.UnsafePointer:
		return nil, nil
	default:
		return nil, nil
	}
}

func applySliceRules(fv reflect.Value, fieldName string, rules []string) (ValidationErrors, error) {
	var errs ValidationErrors
	for i := 0; i < fv.Len(); i++ {
		elemName := fmt.Sprintf("%s[%d]", fieldName, i)
		elemErrs, err := applyValidateRules(fv.Index(i), elemName, rules)
		if err != nil {
			return nil, err
		}
		errs = append(errs, elemErrs...)
	}
	return errs, nil
}

func applyIntRules(n int64, fieldName string, rules []string) (ValidationErrors, error) {
	var errs ValidationErrors

	for _, rule := range rules {
		name, arg, err := parseRule(rule)
		if err != nil {
			return nil, fmt.Errorf("%w in field %q: %w", ErrInvalidValidator, fieldName, err)
		}

		switch name {
		case "min":
			err = validateMin(n, fieldName, arg, &errs)
		case "max":
			err = validateMax(n, fieldName, arg, &errs)
		case "in":
			err = validateIn(n, fieldName, arg, &errs)
		default:
			return nil, fmt.Errorf("%w: unknown validator %q for int field %q", ErrInvalidValidator, name, fieldName)
		}

		if err != nil {
			return nil, err
		}
	}

	return errs, nil
}

func parseIntArg(ruleName, arg, fieldName string) (int64, error) {
	val, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s value %q is not an integer (field %q)", ErrInvalidValidator, ruleName, arg, fieldName)
	}
	return val, nil
}

func validateMin(n int64, fieldName, arg string, errs *ValidationErrors) error {
	minm, err := parseIntArg("min", arg, fieldName)
	if err != nil {
		return err
	}
	if n < minm {
		*errs = append(*errs, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("value %d is less than min %d", n, minm),
		})
	}
	return nil
}

func validateMax(n int64, fieldName, arg string, errs *ValidationErrors) error {
	maxm, err := parseIntArg("max", arg, fieldName)
	if err != nil {
		return err
	}
	if n > maxm {
		*errs = append(*errs, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("value %d is greater than max %d", n, maxm),
		})
	}
	return nil
}

func validateIn(n int64, fieldName, arg string, errs *ValidationErrors) error {
	parts := strings.Split(arg, ",")
	for _, p := range parts {
		v, err := parseIntArg("in", strings.TrimSpace(p), fieldName)
		if err != nil {
			return err
		}
		if n == v {
			return nil
		}
	}
	*errs = append(*errs, ValidationError{
		Field: fieldName,
		Err:   fmt.Errorf("value %d must be one of [%s]", n, arg),
	})
	return nil
}

func parseRule(rule string) (name, arg string, err error) {
	idx := strings.IndexByte(rule, ':')
	if idx < 0 {
		return "", "", fmt.Errorf("rule %q has no colon separator", rule)
	}
	return rule[:idx], rule[idx+1:], nil
}

func applyStringRules(s, fieldName string, rules []string) (ValidationErrors, error) {
	var errs ValidationErrors
	for _, rule := range rules {
		name, arg, err := parseRule(rule)
		if err != nil {
			return nil, fmt.Errorf("%w in field %q: %w", ErrInvalidValidator, fieldName, err)
		}
		switch name {
		case "len":
			n, err := strconv.Atoi(arg)
			if err != nil {
				return nil, fmt.Errorf("%w: len value %q is not an integer (field %q)", ErrInvalidValidator, arg, fieldName)
			}
			if len([]rune(s)) != n {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("length must be %d, got %d", n, len([]rune(s))),
				})
			}
		case "regexp":
			re, err := regexp.Compile(arg)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid regexp %q (field %q): %w", ErrInvalidValidator, arg, fieldName, err)
			}
			if !re.MatchString(s) {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("value %q does not match regexp %q", s, arg),
				})
			}
		case "in":
			allowed := strings.Split(arg, ",")
			if !containsStr(allowed, s) {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("value %q must be one of [%s]", s, arg),
				})
			}
		default:
			return nil, fmt.Errorf("%w: unknown validator %q for string field %q", ErrInvalidValidator, name, fieldName)
		}
	}
	return errs, nil
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
