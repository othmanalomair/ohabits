package ai

// PromptImproveText returns a prompt to improve formatting while keeping the original voice
func PromptImproveText(text string) string {
	return `[مهمة: تنسيق نص عربي بـ Markdown]

نسق هذا النص العربي بصيغة Markdown مع الحفاظ على نفس الكلمات.

التنسيق المطلوب:
- ## للعناوين الرئيسية
- ### للعناوين الفرعية
- **كلمة** للكلمات المهمة فقط (ليس كل النص)
- - أو 1. للقوائم
- --- للفصل بين الأقسام

قواعد:
- لا تغير الكلمات أو الأسلوب
- لا تضف كلام من عندك
- ابدأ مباشرة بالنص المنسق

النص:
` + text + `

النص المنسق:`
}

// PromptFixErrors returns a prompt to fix only spelling errors without changing style
func PromptFixErrors(text string) string {
	return `[مهمة: تصحيح إملائي]

صحح الأخطاء الإملائية فقط في هذا النص العربي.

قواعد:
- صحح الإملاء فقط
- لا تغير الأسلوب أو الكلمات
- لا تضف نجوم أو رموز
- لا تضف أي كلام من عندك

النص:
` + text + `

النص المصحح:`
}

// PromptSimplifyText returns a prompt to organize text without changing the voice
func PromptSimplifyText(text string) string {
	return `[مهمة: ترتيب نص عربي بـ Markdown]

رتب أفكار هذا النص العربي بشكل منطقي مع تنسيق Markdown.

التنسيق المطلوب:
- ## للعناوين الرئيسية
- ### للعناوين الفرعية
- - أو 1. للقوائم
- --- للفصل بين الأقسام

قواعد:
- لا تغير الكلمات أو الأسلوب
- رتب الأفكار المتشابهة مع بعض
- لا تضف كلام من عندك
- ابدأ مباشرة بالنص المرتب

النص:
` + text + `

النص المرتب:`
}

// PromptGenerateTitles returns a prompt to generate blog titles
func PromptGenerateTitles(content string) string {
	// Truncate content if too long (keep first 1000 characters)
	runes := []rune(content)
	if len(runes) > 1000 {
		content = string(runes[:1000]) + "..."
	}

	return `[مهمة: اقتراح عناوين]

اقترح 3 عناوين عربية قصيرة لهذه المدونة.

قواعد:
- عناوين عربية فقط
- كل عنوان 5-10 كلمات
- اكتب كل عنوان في سطر مرقم
- لا تضف شرح

المحتوى:
` + content + `

العناوين:
1. `
}
