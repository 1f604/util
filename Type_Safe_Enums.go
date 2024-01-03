// This is a horrible hack to work around Go's lack of built-in support for type-safe enums
// True type safe enums are not possible in Go due to its lack of sum types
// This is a "best effort" workaround that will prevent some kinds of bugs but is not perfect
package util

type HandlerTypeEnum interface {
	isHandlerEnumTypeValue()
}

type LONGEST_PREFIX_HANDLER_t struct{}
type EXACT_MATCH_HANDLER_t struct{}

func (LONGEST_PREFIX_HANDLER_t) isHandlerEnumTypeValue() {}
func (EXACT_MATCH_HANDLER_t) isHandlerEnumTypeValue()    {}

var EXACT_MATCH_HANDLER EXACT_MATCH_HANDLER_t = EXACT_MATCH_HANDLER_t{}
var LONGEST_PREFIX_HANDLER LONGEST_PREFIX_HANDLER_t = LONGEST_PREFIX_HANDLER_t{}

/* Usage:

switch val.(type) {
case PREFIX_MATCH_HANDLER:
    …
case EXACT_MATCH_HANDLER:
    …
default:
    // nil comes here
    return fmt.Errorf("unsupported num value %T", val)
}
*/

type MapItemValueType interface {
	isMapItemEnumTypeValue()
	ToString() string
}

type URL_TYPE_t struct{}
type PASTE_TYPE_t struct{}

func (URL_TYPE_t) isMapItemEnumTypeValue()   {}
func (PASTE_TYPE_t) isMapItemEnumTypeValue() {}
func (URL_TYPE_t) ToString() string {
	return "url"
}
func (PASTE_TYPE_t) ToString() string {
	return "paste"
}

var TYPE_MAP_ITEM_URL URL_TYPE_t = URL_TYPE_t{}
var TYPE_MAP_ITEM_PASTE PASTE_TYPE_t = PASTE_TYPE_t{}
