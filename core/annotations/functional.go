package annotations

func GetDescription(holder *AnnotationHolder) string {
	if holder == nil {
		return ""
	}
	return holder.GetDescription()
}
