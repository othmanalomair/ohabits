package ai

// PromptImproveText returns a prompt to improve formatting while keeping the original voice
func PromptImproveText(text string) string {
	return `[مهمة: تحسين النص بـ Markdown مع الحفاظ على اللهجة]

حسّن تنسيق النص بصيغة Markdown مع الحفاظ على اللهجة العامية.

المطلوب:
1. تنسيق Markdown:
   - أضف # للعناوين الرئيسية و ## للفرعية إذا مو موجودة
   - أضف **نص** للكلمات المهمة
   - أضف - للقوائم إذا فيه تعداد
   - رتب الفقرات بشكل منطقي

2. تصحيح بسيط:
   - صحح الكلمات الإنجليزية (apple → Apple، macbook → MacBook)
   - صحح الأخطاء الإملائية الواضحة فقط

مهم جداً - الحفاظ على اللهجة:
- "شريت" تبقى "شريت" (مو "اشتريت")
- "وايد" تبقى "وايد" (مو "كثير")
- "الحين" تبقى "الحين" (مو "الآن")
- "شنو" تبقى "شنو" (مو "ماذا")
- "ابي/ابيه" تبقى "ابي/ابيه" (مو "أريد")

ممنوع:
- تحويل العامية إلى فصحى
- تغيير أسلوب الكاتب
- حذف أو إضافة معلومات جديدة

النص:
` + text + `

النص المحسّن:`
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

// PromptCustomEdit returns a prompt to apply custom user instructions to text
func PromptCustomEdit(text string, userInstructions string) string {
	return `[مهمة: تعديل نص حسب التعليمات]

طبّق التعليمات التالية على النص العربي أدناه.

التعليمات:
` + userInstructions + `

قواعد:
- نفذ التعليمات بدقة
- حافظ على معنى النص الأساسي
- لا تضف معلومات غير مطلوبة
- ابدأ مباشرة بالنص المعدل

النص الأصلي:
` + text + `

النص المعدل:`
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

// PromptMonthlySummary returns a prompt to generate a monthly summary from notes
func PromptMonthlySummary(monthName string, year int, notesContent string) string {
	// Truncate content if too long (keep first 4000 characters)
	runes := []rune(notesContent)
	if len(runes) > 4000 {
		notesContent = string(runes[:4000]) + "..."
	}

	return `[مهمة: ملخص شهري]

اكتب ملخصاً موجزاً لشهر ` + monthName + ` ` + formatYear(year) + ` بناءً على المذكرات اليومية التالية.

قواعد:
- ملخص عربي موجز (3-5 جمل)
- ركز على الأحداث المهمة والإنجازات
- اذكر المشاعر السائدة إن وجدت
- لا تكرر تفاصيل كل يوم
- ابدأ مباشرة بالملخص بدون مقدمة
- لا تضف عناوين أو ترقيم

المذكرات:
` + notesContent + `

الملخص:`
}

func formatYear(year int) string {
	return string(rune('0'+year/1000)) + string(rune('0'+(year/100)%10)) + string(rune('0'+(year/10)%10)) + string(rune('0'+year%10))
}
