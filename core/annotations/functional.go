package annotations

func GetDescription(holder *AnnotationHolder) string {
	if holder == nil {
		return ""
	}
	return holder.GetDescription()
}

func GetFirstValueOrEmpty(holder *AnnotationHolder, annotation GleeceAnnotation) string {
	if holder == nil {
		return ""
	}
	return holder.GetFirstValueOrEmpty(annotation)
}
