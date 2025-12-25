package keyboard

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AddBackButton добавляет кнопку "назад" (⬅️) в правый нижний угол клавиатуры
// Возвращает новую клавиатуру с добавленной кнопкой "назад"
func AddBackButton(keyboard tgbotapi.InlineKeyboardMarkup) tgbotapi.InlineKeyboardMarkup {
	// Создаём кнопку "назад"
	btnBack := tgbotapi.NewInlineKeyboardButtonData("⬅️", "nav_back")

	// Добавляем кнопку "назад" в отдельный ряд (правый нижний угол)
	backRow := tgbotapi.NewInlineKeyboardRow(btnBack)

	// Добавляем новый ряд к существующим рядам
	rows := append(keyboard.InlineKeyboard, backRow)
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return newKeyboard
}

// NewConfirmKeyboard создаёт клавиатуру с кнопками "Да" и "Нет"
func NewConfirmKeyboard(dataPrefix string) tgbotapi.InlineKeyboardMarkup {
	// Создаём инлайн-кнопки
	btnYes := tgbotapi.NewInlineKeyboardButtonData("✅ Да", dataPrefix+"_yes")
	btnNo := tgbotapi.NewInlineKeyboardButtonData("❌ Нет", dataPrefix+"_no")

	// Создаём ряд кнопок
	row := tgbotapi.NewInlineKeyboardRow(btnYes, btnNo)

	// Создаём клавиатуру
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	return keyboard
}

// NewNotificationKeyboard создаёт клавиатуру для переключения уведомлений
// enabled - текущее состояние уведомлений (true = включено, false = выключено)
func NewNotificationKeyboard(enabled bool) tgbotapi.InlineKeyboardMarkup {
	var btnText string
	var callbackData string

	if enabled {
		btnText = "🔔 Уведомления: Вкл"
		callbackData = "notif_off" // При нажатии переключим на выключено
	} else {
		btnText = "🔕 Уведомления: Выкл"
		callbackData = "notif_on" // При нажатии переключим на включено
	}

	btn := tgbotapi.NewInlineKeyboardButtonData(btnText, callbackData)
	row := tgbotapi.NewInlineKeyboardRow(btn)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	return keyboard
}

// NewLanguageInlineKeyboard создаёт inline-клавиатуру для выбора языка интерфейса
// currentLang - текущий выбранный язык (ru, en, zh)
func NewLanguageInlineKeyboard(currentLang string) tgbotapi.InlineKeyboardMarkup {
	// Создаём кнопки для выбора языка
	btnRu := tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru")
	btnEn := tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "lang_en")
	btnZh := tgbotapi.NewInlineKeyboardButtonData("🇨🇳 中文", "lang_zh")

	// Добавляем галочку к текущему выбранному языку
	switch currentLang {
	case "ru":
		btnRu = tgbotapi.NewInlineKeyboardButtonData("✅ 🇷🇺 Русский", "lang_ru")
	case "en":
		btnEn = tgbotapi.NewInlineKeyboardButtonData("✅ 🇬🇧 English", "lang_en")
	case "zh":
		btnZh = tgbotapi.NewInlineKeyboardButtonData("✅ 🇨🇳 中文", "lang_zh")
	}

	// Размещаем кнопки в один ряд
	row := tgbotapi.NewInlineKeyboardRow(btnRu, btnEn, btnZh)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	return keyboard
}

// NewMainMenuInlineKeyboard создаёт inline-клавиатуру для главного меню
// С кнопками: Профиль, Настройки, Меню, Курсы
func NewMainMenuInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	btnProfile := tgbotapi.NewInlineKeyboardButtonData("👤 Профиль", "menu_profile")
	btnSettings := tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "menu_settings")
	btnMenu := tgbotapi.NewInlineKeyboardButtonData("📋 Меню", "menu_menu")
	btnCourses := tgbotapi.NewInlineKeyboardButtonData("📚 Курсы", "menu_courses")

	// Первый ряд - Профиль и Настройки
	row1 := tgbotapi.NewInlineKeyboardRow(btnProfile, btnSettings)
	// Второй ряд - Меню и Курсы
	row2 := tgbotapi.NewInlineKeyboardRow(btnMenu, btnCourses)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
	return keyboard
}

// Course представляет курс
type Course struct {
	ID          int
	Title       string
	Description string
}

// NewCoursesKeyboard создаёт inline-клавиатуру для пагинации курсов
// courses - список всех курсов
// currentPage - текущая страница (начинается с 0)
// itemsPerPage - количество курсов на странице
func NewCoursesKeyboard(courses []Course, currentPage, itemsPerPage int) tgbotapi.InlineKeyboardMarkup {
	totalPages := (len(courses) + itemsPerPage - 1) / itemsPerPage // Округление вверх
	if totalPages == 0 {
		totalPages = 1
	}

	// Ограничиваем currentPage в допустимых пределах
	if currentPage < 0 {
		currentPage = 0
	}
	if currentPage >= totalPages {
		currentPage = totalPages - 1
	}

	// Вычисляем индексы для текущей страницы
	startIdx := currentPage * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if endIdx > len(courses) {
		endIdx = len(courses)
	}

	// Создаём кнопки для курсов на текущей странице
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := startIdx; i < endIdx; i++ {
		course := courses[i]
		btnText := fmt.Sprintf("%d. %s", i+1, course.Title)
		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("course_%d", course.ID))
		row := tgbotapi.NewInlineKeyboardRow(btn)
		rows = append(rows, row)
	}

	// Создаём кнопки навигации (⬅️ Назад / Вперёд ➡️)
	var navRow []tgbotapi.InlineKeyboardButton

	// Кнопка "Назад" (⬅️)
	if currentPage > 0 {
		btnPrev := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("courses_page_%d", currentPage-1))
		navRow = append(navRow, btnPrev)
	}

	// Информация о странице (если есть несколько страниц)
	if totalPages > 1 {
		pageInfo := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d/%d", currentPage+1, totalPages),
			"courses_info",
		)
		navRow = append(navRow, pageInfo)
	}

	// Кнопка "Вперёд" (➡️)
	if currentPage < totalPages-1 {
		btnNext := tgbotapi.NewInlineKeyboardButtonData("Вперёд ➡️", fmt.Sprintf("courses_page_%d", currentPage+1))
		navRow = append(navRow, btnNext)
	}

	// Добавляем ряд навигации, если есть кнопки
	if len(navRow) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(navRow...))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return keyboard
}
