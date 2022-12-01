package prago

import "fmt"

type FormItemAction[T any] struct {
	data *formItemActionData
}

type formItemActionData struct {
	resourceData *resourceData
	formAction   *FormAction
}

func (resource *Resource[T]) FormItemAction(url string) *FormItemAction[T] {
	return &FormItemAction[T]{
		data: resource.data.FormItemAction(url),
	}
}

func (resourceData *resourceData) FormItemAction(url string) *formItemActionData {
	fa := newFormAction(resourceData.app, url, func(f *Form, r *Request) {
		fmt.Println("HEEER")
		fmt.Println(r.Params())
		fmt.Println(r.Param("id"))
		item := resourceData.query().ID(r.Param("id"))
		f.image = resourceData.app.getItemImageLarge(item)
	})

	fa.actionForm.resourceData = resourceData
	fa.actionValidation.resourceData = resourceData

	fa.actionForm.Permission(resourceData.canView)
	fa.actionValidation.Permission(resourceData.canView)

	fa.actionForm.isItemAction = true
	fa.actionValidation.isItemAction = true

	resourceData.itemActions = append(resourceData.itemActions, fa.actionForm)
	resourceData.itemActions = append(resourceData.itemActions, fa.actionValidation)
	return &formItemActionData{
		resourceData: resourceData,
		formAction:   fa,
	}

}

func (actionData *formItemActionData) priority() *formItemActionData {
	actionData.formAction.priority()
	return actionData
}

func (action *FormItemAction[T]) Permission(permission Permission) *FormItemAction[T] {
	action.data.Permission(permission)
	return action
}

func (actionData *formItemActionData) Permission(permission Permission) *formItemActionData {
	actionData.formAction.Permission(permission)
	return actionData
}

func (action *FormItemAction[T]) Icon(icon string) *FormItemAction[T] {
	action.data.Icon(icon)
	return action
}

func (actionData *formItemActionData) Icon(icon string) *formItemActionData {
	actionData.formAction.Icon(icon)
	return actionData
}

func (action *FormItemAction[T]) Name(name func(string) string) *FormItemAction[T] {
	action.data.Name(name)
	return action
}

func (actionData *formItemActionData) Name(name func(string) string) *formItemActionData {
	actionData.formAction.Name(name)
	return actionData
}

func (action *FormItemAction[T]) Form(formGenerator func(*T, *Form, *Request)) *FormItemAction[T] {
	action.data.Form(func(a any, f *Form, r *Request) {
		formGenerator(a.(*T), f, r)
	})
	return action
}

func (actionData *formItemActionData) Form(formGenerator func(any, *Form, *Request)) *formItemActionData {
	actionData.formAction.Form(func(form *Form, request *Request) {
		item := actionData.resourceData.ID(request.Param("id"))
		if item == nil {
			render404(request)
			return
		}
		formGenerator(item, form, request)
	})
	return actionData
}

func (action *FormItemAction[T]) Validation(validation func(*T, ValidationContext)) *FormItemAction[T] {
	action.data.Validation(func(a any, vc ValidationContext) {
		validation(a.(*T), vc)
	})
	return action
}

func (actionData *formItemActionData) Validation(validation func(any, ValidationContext)) *formItemActionData {
	actionData.formAction.Validation(func(vc ValidationContext) {
		item := actionData.resourceData.ID(vc.GetValue("id"))
		if item == nil {
			panic("can't find item")
		}
		validation(item, vc)
	})
	return actionData
}
