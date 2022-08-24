package prago

type FormItemAction[T any] struct {
	resource   *Resource[T]
	formAction *FormAction
}

func (resource *Resource[T]) FormItemAction(url string) *FormItemAction[T] {
	fa := newFormAction(resource.data.app, url)

	fa.actionForm.resourceData = resource.data
	fa.actionValidation.resourceData = resource.data

	fa.actionForm.Permission(resource.data.canView)
	fa.actionValidation.Permission(resource.data.canView)

	fa.actionForm.isItemAction = true
	fa.actionValidation.isItemAction = true

	resource.data.itemActions = append(resource.data.itemActions, fa.actionForm)
	resource.data.itemActions = append(resource.data.itemActions, fa.actionValidation)
	return &FormItemAction[T]{
		resource:   resource,
		formAction: fa,
	}
}

func (action *FormItemAction[T]) priority() *FormItemAction[T] {
	action.formAction.priority()
	return action
}

func (action *FormItemAction[T]) Permission(permission Permission) *FormItemAction[T] {
	action.formAction.Permission(permission)
	return action
}

func (action *FormItemAction[T]) Name(name func(string) string) *FormItemAction[T] {
	action.formAction.Name(name)
	return action
}

func (action *FormItemAction[T]) Form(formGenerator func(*T, *Form, *Request)) *FormItemAction[T] {
	action.formAction.Form(func(form *Form, request *Request) {
		item := action.resource.ID(request.Param("id"))
		if item == nil {
			render404(request)
			return
		}
		formGenerator(item, form, request)
	})
	return action
}

func (action *FormItemAction[T]) Validation(validation func(*T, ValidationContext)) *FormItemAction[T] {
	action.formAction.Validation(func(vc ValidationContext) {
		item := action.resource.ID(vc.GetValue("id"))
		if item == nil {
			panic("can't find item")
		}
		validation(item, vc)
	})
	return action
}
