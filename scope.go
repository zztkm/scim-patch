package scimpatch

import (
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/schema"
)

type ScopeNavigator struct {
	op   scim.PatchOperation
	data scim.ResourceAttributes
	attr schema.CoreAttribute
}

func NewScopeNavigator(op scim.PatchOperation, data scim.ResourceAttributes, attr schema.CoreAttribute) *ScopeNavigator {
	return &ScopeNavigator{
		op:   op,
		data: data,
		attr: attr,
	}
}

// GetMap は 処理対象であるmapまでのスコープをたどり該当のmapを返却します
func (n *ScopeNavigator) GetMap() scim.ResourceAttributes {
	return n.data
}

// GetScopedMap は 処理対象であるmapまでのスコープをたどり該当のmapを返却します
func (n *ScopeNavigator) GetScopedMap() (scim.ResourceAttributes, string) {
	return n.getAttributeScopedMap()
}

// ApplyScopedMap は 処理対象であるmapまでのスコープをたどりscopedMapに置換します
func (n *ScopeNavigator) ApplyScopedMap(scopedMap scim.ResourceAttributes) {
	uriScoped := n.getURIScopedMap()
	if _, required := n.requiredSubAttributes(); required {
		uriScoped = n.attatchToMap(uriScoped, scopedMap, n.attr.Name(), required)
	}

	data := n.data
	uriPrefix, containsURI := n.containsURIPrefix()
	data = n.attatchToMap(data, uriScoped, uriPrefix, containsURI)
	n.data = data
}

// getURIScopedMap は URIに応じて、処理対象のMapを返却します
func (n *ScopeNavigator) getURIScopedMap() scim.ResourceAttributes {
	uriScoped := n.data
	uriPrefix, ok := n.containsURIPrefix()
	uriScoped = n.navigateToMap(uriScoped, uriPrefix, ok)
	return uriScoped
}

// getAttributeScopedMap は 属性に応じて、処理対象のMapを返却します
func (n *ScopeNavigator) getAttributeScopedMap() (scim.ResourceAttributes, string) {
	// initialize returns
	data := n.getURIScopedMap()
	subAttrName, ok := n.requiredSubAttributes()
	data = n.navigateToMap(data, n.attr.Name(), ok)
	return data, subAttrName
}

// navigateToMap は必要に応じて、パスをたどる処理です
func (n *ScopeNavigator) navigateToMap(data map[string]interface{}, attr string, ok bool) scim.ResourceAttributes {
	if ok {
		data_, ok := data[attr].(map[string]interface{})
		switch ok {
		case true:
			data = data_
		case false:
			data = scim.ResourceAttributes{}
		}
	}
	return data
}

// attatchToMap は必要に応じて、パスを戻す処理です
func (n *ScopeNavigator) attatchToMap(data map[string]interface{}, scoped map[string]interface{}, attr string, ok bool) scim.ResourceAttributes {
	if ok {
		if len(scoped) == 0 {
			delete(data, attr)
		} else {
			data[attr] = scoped
		}
	}
	return data
}

// containsURIPrefix は対象の属性がURIPrefixを持ったmapの中に格納されているかどうかを判断します
func (n *ScopeNavigator) containsURIPrefix() (string, bool) {
	ok := false
	uriPrefix := ""
	if n.op.Path != nil && n.op.Path.AttributePath.URIPrefix != nil {
		ok = true
		uriPrefix = *n.op.Path.AttributePath.URIPrefix
	}
	return uriPrefix, ok
}

// requiredSubAttributes は対象の属性がサブ属性を保持したマップであるかどうかと、サブ属性が対象となったPatchOpeartionかどうかを判断します
func (n *ScopeNavigator) requiredSubAttributes() (string, bool) {
	ok := false
	subAttr := n.attr.Name()
	if n.attr.HasSubAttributes() && n.op.Path != nil && n.op.Path.AttributePath.SubAttribute != nil {
		ok = true
		subAttr = *n.op.Path.AttributePath.SubAttribute
	}
	return subAttr, ok
}
