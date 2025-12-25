package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/config"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/keyboard"
	"telegram-bot/internal/middleware"
)

// userNotifications хранит состояние уведомлений для каждого пользователя
// Ключ - chatID, значение - включены ли уведомления
var (
	userNotifications = make(map[int64]bool)
	notificationsMu   sync.RWMutex // Мьютекс для безопасного доступа к map
)

// userLanguages хранит выбранный язык для каждого пользователя
// Ключ - chatID, значение - код языка (ru, en, zh)
var (
	userLanguages = make(map[int64]string)
	languagesMu   sync.RWMutex // Мьютекс для безопасного доступа к map
)

// NavigationState хранит состояние навигации для возврата назад
type NavigationState struct {
	Text      string                         // Текст сообщения
	Keyboard  *tgbotapi.InlineKeyboardMarkup // Клавиатура
	MessageID int                            // ID сообщения (для редактирования)
}

// userNavigationHistory хранит историю навигации для каждого пользователя
// Ключ - chatID, значение - стек состояний (последний элемент - текущее состояние)
var (
	userNavigationHistory = make(map[int64][]NavigationState)
	navigationMu          sync.RWMutex // Мьютекс для безопасного доступа к map
)

// coursesList содержит список всех курсов
var coursesList = []keyboard.Course{
	{ID: 1, Title: "Go для начинающих", Description: "Изучите основы языка Go"},
	{ID: 2, Title: "Продвинутый Go", Description: "Углублённое изучение Go"},
	{ID: 3, Title: "Telegram Bot API", Description: "Создание ботов на Go"},
	{ID: 4, Title: "Базы данных в Go", Description: "Работа с PostgreSQL и MySQL"},
	{ID: 5, Title: "Микросервисы на Go", Description: "Архитектура микросервисов"},
	{ID: 6, Title: "Тестирование в Go", Description: "Unit и интеграционные тесты"},
	{ID: 7, Title: "Конкурентность в Go", Description: "Goroutines и Channels"},
	{ID: 8, Title: "REST API на Go", Description: "Создание RESTful API"},
	{ID: 9, Title: "Docker и Go", Description: "Контейнеризация приложений"},
	{ID: 10, Title: "Deployment Go приложений", Description: "Развёртывание на сервере"},
}

// userCoursesPage хранит текущую страницу курсов для каждого пользователя
// Ключ - chatID, значение - номер страницы (начинается с 0)
var (
	userCoursesPage = make(map[int64]int)
	coursesPageMu   sync.RWMutex // Мьютекс для безопасного доступа к map
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Создаём экземпляр бота
	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	bot.Debug = cfg.Bot.Debug
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Создаём диспетчер обработчиков
	dispatcher := handler.NewDispatcher()

	// Регистрируем обработчики команд
	dispatcher.Register(handler.NewStartHandler())
	dispatcher.Register(handler.NewHelpHandler())
	dispatcher.Register(handler.NewInfoHandler())
	dispatcher.Register(handler.NewAdminHandler(cfg.Bot.AdminIDs))

	// Создаём обработчик обычных сообщений
	messageHandler := handler.NewMessageHandler()

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Bot.Timeout
	updates := bot.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		handleUpdate(bot, dispatcher, messageHandler, update)
	}
}

func handleUpdate(
	bot *tgbotapi.BotAPI,
	dispatcher *handler.Dispatcher,
	messageHandler *handler.MessageHandler,
	update tgbotapi.Update,
) {
	// Обрабатываем callback-запросы (нажатия на инлайн-кнопки)
	if update.CallbackQuery != nil {
		handleCallbackQuery(bot, update.CallbackQuery)
		return
	}

	// Обрабатываем сообщения
	if update.Message == nil {
		return
	}

	msg := update.Message

	if msg.IsCommand() {
		middleware.LogCommand(msg)
		err := dispatcher.HandleCommand(bot, msg)
		if err != nil {
			log.Printf("Ошибка обработки команды: %v", err)
		}
		return
	}

	if msg.Text != "" {
		middleware.LogMessage(msg)
		// Передаём функции получения состояния уведомлений и языка
		err := messageHandler.Handle(bot, msg, getNotificationState, getLanguage)
		if err != nil {
			log.Printf("Ошибка обработки сообщения: %v", err)
		}
	}
}

// handleCallbackQuery обрабатывает нажатие на инлайн-кнопку
func handleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID
	userID := callback.From.ID

	log.Printf("Callback от пользователя %d: %s", userID, data)

	// Отвечаем на callback-запрос (обязательно!)
	// Сначала отвечаем пустым ответом, затем при необходимости обновим с текстом
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(callbackConfig); err != nil {
		log.Printf("Ошибка ответа на callback: %v", err)
		return
	}

	// Сохраняем текущее состояние перед переходом (если это не кнопка "назад")
	if data != "nav_back" {
		currentText := callback.Message.Text
		// Получаем текущую клавиатуру из сообщения
		currentKeyboard := callback.Message.ReplyMarkup
		saveNavigationState(chatID, currentText, currentKeyboard, messageID)
	}

	// Обрабатываем данные в зависимости от префикса
	switch {
	case strings.HasPrefix(data, "delete_profile_"):
		handleDeleteProfile(bot, callback.ID, chatID, messageID, data)

	case strings.HasPrefix(data, "notif_"):
		handleNotificationToggle(bot, callback.ID, chatID, messageID, data)

	case strings.HasPrefix(data, "lang_"):
		handleLanguageChange(bot, callback.ID, chatID, messageID, data)

	case strings.HasPrefix(data, "settings_"):
		handleSettingsMenu(bot, chatID, messageID, data)

	case strings.HasPrefix(data, "menu_"):
		handleMainMenuNavigation(bot, chatID, messageID, data)

	case strings.HasPrefix(data, "courses_"):
		handleCoursesNavigation(bot, callback.ID, chatID, messageID, data)

	case strings.HasPrefix(data, "course_"):
		handleCourseDetails(bot, callback.ID, chatID, messageID, data)

	case data == "nav_back":
		handleBackNavigation(bot, chatID, messageID)

	default:
		// Обработка неизвестных callback-запросов
		log.Printf("Неизвестный callback-запрос: %s от пользователя %d", data, userID)
		// Отвечаем с сообщением об ошибке
		callbackConfig := tgbotapi.NewCallback(callback.ID, "❌ Неизвестная команда")
		bot.Request(callbackConfig)
	}
}

// handleMainMenuNavigation обрабатывает навигацию из главного меню
func handleMainMenuNavigation(bot *tgbotapi.BotAPI, chatID int64, messageID int, data string) {
	switch data {
	case "menu_profile":
		// Переход к профилю
		handleMenuProfile(bot, chatID, messageID)
	case "menu_settings":
		// Переход к настройкам
		handleMenuSettings(bot, chatID, messageID)
	case "menu_courses":
		// Переход к курсам
		handleMenuCourses(bot, chatID, messageID)
	case "menu_menu":
		// Уже в меню - ничего не делаем или обновляем сообщение
		text := "📋 Главное меню:\n\n" +
			"Доступные разделы:\n" +
			"• Профиль - информация о вас\n" +
			"• Настройки - настройки бота\n" +
			"• Меню - это сообщение\n" +
			"• Курсы - список доступных курсов"
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		kb := keyboard.NewMainMenuInlineKeyboard()
		kb = keyboard.AddBackButton(kb)
		edit.ReplyMarkup = &kb
		bot.Send(edit)
	}
}

// handleMenuProfile обрабатывает переход к профилю из главного меню
func handleMenuProfile(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	user := bot.Self
	text := fmt.Sprintf("👤 Ваш профиль:\n\n"+
		"ID: %d\n"+
		"Имя: %s\n"+
		"Username: @%s\n\n"+
		"Хотите удалить профиль?", chatID, user.FirstName, user.UserName)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	kb := keyboard.NewConfirmKeyboard("delete_profile")
	kb = keyboard.AddBackButton(kb)
	edit.ReplyMarkup = &kb
	bot.Send(edit)
}

// handleMenuSettings обрабатывает переход к настройкам из главного меню
func handleMenuSettings(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := "⚙️ Настройки:\n\n" +
		"Выберите настройку для изменения:"

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	btnNotif := tgbotapi.NewInlineKeyboardButtonData("🔔 Уведомления", "settings_notif")
	btnLang := tgbotapi.NewInlineKeyboardButtonData("🌐 Язык", "settings_lang")
	row := tgbotapi.NewInlineKeyboardRow(btnNotif, btnLang)
	kb := tgbotapi.NewInlineKeyboardMarkup(row)
	kb = keyboard.AddBackButton(kb)
	edit.ReplyMarkup = &kb
	bot.Send(edit)
}

// handleMenuCourses обрабатывает переход к курсам из главного меню
func handleMenuCourses(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	// Получаем текущую страницу пользователя (по умолчанию 0)
	coursesPageMu.RLock()
	currentPage := userCoursesPage[chatID]
	coursesPageMu.RUnlock()

	// Показываем курсы на текущей странице
	showCoursesPage(bot, chatID, messageID, currentPage)
}

// showCoursesPage показывает курсы на указанной странице
func showCoursesPage(bot *tgbotapi.BotAPI, chatID int64, messageID int, page int) {
	const itemsPerPage = 3 // Количество курсов на странице

	// Обновляем текущую страницу пользователя
	coursesPageMu.Lock()
	userCoursesPage[chatID] = page
	coursesPageMu.Unlock()

	// Вычисляем индексы для текущей страницы
	startIdx := page * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if endIdx > len(coursesList) {
		endIdx = len(coursesList)
	}

	// Формируем текст с курсами на текущей странице
	text := "📚 Доступные курсы:\n\n"
	for i := startIdx; i < endIdx; i++ {
		course := coursesList[i]
		text += fmt.Sprintf("%d. %s\n%s\n\n", i+1, course.Title, course.Description)
	}

	// Создаём клавиатуру с пагинацией
	kb := keyboard.NewCoursesKeyboard(coursesList, page, itemsPerPage)
	kb = keyboard.AddBackButton(kb)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ReplyMarkup = &kb
	bot.Send(edit)
}

// handleCoursesNavigation обрабатывает навигацию по страницам курсов
func handleCoursesNavigation(bot *tgbotapi.BotAPI, callbackID string, chatID int64, messageID int, data string) {
	if data == "courses_info" {
		// Просто обновляем текущую страницу (информация о странице)
		coursesPageMu.RLock()
		currentPage := userCoursesPage[chatID]
		coursesPageMu.RUnlock()
		showCoursesPage(bot, chatID, messageID, currentPage)
		return
	}

	// Извлекаем номер страницы из callback data (формат: "courses_page_0")
	if strings.HasPrefix(data, "courses_page_") {
		var page int
		_, err := fmt.Sscanf(data, "courses_page_%d", &page)
		if err != nil {
			log.Printf("Ошибка парсинга номера страницы: %v", err)
			// Отвечаем на callback с ошибкой
			callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Ошибка навигации")
			bot.Request(callbackConfig)
			return
		}
		showCoursesPage(bot, chatID, messageID, page)
	} else {
		// Неизвестный callback - отвечаем с ошибкой
		callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Неизвестная команда")
		bot.Request(callbackConfig)
	}
}

// handleCourseDetails обрабатывает нажатие на конкретный курс
func handleCourseDetails(bot *tgbotapi.BotAPI, callbackID string, chatID int64, messageID int, data string) {
	// Извлекаем ID курса из callback data (формат: "course_1")
	var courseID int
	_, err := fmt.Sscanf(data, "course_%d", &courseID)
	if err != nil {
		log.Printf("Ошибка парсинга ID курса: %v", err)
		// Отвечаем на callback с ошибкой
		callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Ошибка загрузки курса")
		bot.Request(callbackConfig)
		return
	}

	// Находим курс по ID
	var course *keyboard.Course
	for i := range coursesList {
		if coursesList[i].ID == courseID {
			course = &coursesList[i]
			break
		}
	}

	if course == nil {
		text := "❌ Курс не найден"
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		// Отвечаем на callback с ошибкой
		callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Курс не найден")
		bot.Request(callbackConfig)
		if _, err := bot.Send(edit); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		return
	}

	// Отвечаем на callback (успешная загрузка)
	callbackConfig := tgbotapi.NewCallback(callbackID, "")
	bot.Request(callbackConfig)

	// Показываем детали курса
	text := fmt.Sprintf("📚 %s\n\n%s", course.Title, course.Description)

	// Кнопка "Назад к списку курсов"
	btnBack := tgbotapi.NewInlineKeyboardButtonData("⬅️ К списку курсов", "menu_courses")
	row := tgbotapi.NewInlineKeyboardRow(btnBack)
	kb := tgbotapi.NewInlineKeyboardMarkup(row)
	kb = keyboard.AddBackButton(kb)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ReplyMarkup = &kb
	if _, err := bot.Send(edit); err != nil {
		log.Printf("Ошибка обновления сообщения: %v", err)
	}
}

// handleBackNavigation обрабатывает нажатие на кнопку "назад"
func handleBackNavigation(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	navigationMu.Lock()
	defer navigationMu.Unlock()

	history, exists := userNavigationHistory[chatID]
	if !exists || len(history) == 0 {
		// Нет истории - возвращаемся в главное меню
		text := "📋 Главное меню:\n\n" +
			"Доступные разделы:\n" +
			"• Профиль - информация о вас\n" +
			"• Настройки - настройки бота\n" +
			"• Меню - это сообщение\n" +
			"• Курсы - список доступных курсов"
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		kb := keyboard.NewMainMenuInlineKeyboard()
		kb = keyboard.AddBackButton(kb)
		edit.ReplyMarkup = &kb
		bot.Send(edit)
		return
	}

	// Убираем текущее состояние из истории
	history = history[:len(history)-1]

	// Если есть предыдущее состояние - восстанавливаем его
	if len(history) > 0 {
		prevState := history[len(history)-1]
		edit := tgbotapi.NewEditMessageText(chatID, messageID, prevState.Text)
		if prevState.Keyboard != nil {
			edit.ReplyMarkup = prevState.Keyboard
		}
		bot.Send(edit)
		userNavigationHistory[chatID] = history
	} else {
		// История пуста - возвращаемся в главное меню
		text := "📋 Главное меню:\n\n" +
			"Доступные разделы:\n" +
			"• Профиль - информация о вас\n" +
			"• Настройки - настройки бота\n" +
			"• Меню - это сообщение\n" +
			"• Курсы - список доступных курсов"
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		kb := keyboard.NewMainMenuInlineKeyboard()
		kb = keyboard.AddBackButton(kb)
		edit.ReplyMarkup = &kb
		bot.Send(edit)
		delete(userNavigationHistory, chatID)
	}
}

// saveNavigationState сохраняет текущее состояние в историю навигации
func saveNavigationState(chatID int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup, messageID int) {
	navigationMu.Lock()
	defer navigationMu.Unlock()

	state := NavigationState{
		Text:      text,
		Keyboard:  keyboard,
		MessageID: messageID,
	}

	history, exists := userNavigationHistory[chatID]
	if !exists {
		history = []NavigationState{}
	}

	// Добавляем текущее состояние в историю
	history = append(history, state)
	userNavigationHistory[chatID] = history
}

// handleSettingsMenu обрабатывает выбор настройки из меню
func handleSettingsMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int, data string) {
	if data == "settings_notif" {
		// Показываем настройки уведомлений
		notificationsEnabled := getNotificationState(chatID)
		var stateText string
		if notificationsEnabled {
			stateText = "Вкл"
		} else {
			stateText = "Выкл"
		}

		editText := "⚙️ Настройки уведомлений:\n\n" +
			"Текущее состояние: " + stateText + "\n\n" +
			"Нажмите кнопку, чтобы переключить:"
		edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
		kb := keyboard.NewNotificationKeyboard(notificationsEnabled)
		kb = keyboard.AddBackButton(kb)
		edit.ReplyMarkup = &kb
		bot.Send(edit)
	} else if data == "settings_lang" {
		// Показываем выбор языка
		currentLang := getLanguage(chatID)
		editText := "🌐 Выбор языка интерфейса:\n\n" +
			"Выберите язык:"
		edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
		kb := keyboard.NewLanguageInlineKeyboard(currentLang)
		kb = keyboard.AddBackButton(kb)
		edit.ReplyMarkup = &kb
		bot.Send(edit)
	}
}

// handleNotificationToggle обрабатывает переключение уведомлений
func handleNotificationToggle(bot *tgbotapi.BotAPI, callbackID string, chatID int64, messageID int, data string) {
	var newState bool
	var statusText string
	var callbackText string

	if data == "notif_on" {
		// Включаем уведомления
		newState = true
		statusText = "🔔 Уведомления включены!"
		callbackText = "✅ Уведомления включены"
	} else if data == "notif_off" {
		// Выключаем уведомления
		newState = false
		statusText = "🔕 Уведомления выключены!"
		callbackText = "✅ Уведомления выключены"
	} else {
		// Неизвестный callback - отвечаем с ошибкой
		callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Неизвестная команда")
		bot.Request(callbackConfig)
		return
	}

	// Сохраняем новое состояние в памяти
	notificationsMu.Lock()
	userNotifications[chatID] = newState
	notificationsMu.Unlock()

	// Отвечаем на callback с текстом (покажет уведомление пользователю)
	callbackConfig := tgbotapi.NewCallback(callbackID, callbackText)
	bot.Request(callbackConfig)

	// Обновляем сообщение с новым состоянием и клавиатурой
	editText := "⚙️ Настройки уведомлений:\n\n" + statusText + "\n\nНажмите кнопку, чтобы переключить:"
	edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
	kb := keyboard.NewNotificationKeyboard(newState)
	kb = keyboard.AddBackButton(kb)
	edit.ReplyMarkup = &kb
	if _, err := bot.Send(edit); err != nil {
		log.Printf("Ошибка обновления сообщения: %v", err)
	}
}

// handleDeleteProfile обрабатывает подтверждение удаления профиля
func handleDeleteProfile(bot *tgbotapi.BotAPI, callbackID string, chatID int64, messageID int, data string) {
	var editText string
	var callbackText string

	if data == "delete_profile_yes" {
		// Пользователь подтвердил удаление
		editText = "✅ Профиль удалён!\n\n" +
			"Все ваши данные были удалены из системы."
		callbackText = "✅ Профиль удалён"
	} else if data == "delete_profile_no" {
		// Пользователь отменил удаление
		editText = "❌ Удаление отменено.\n\n" +
			"Ваш профиль сохранён."
		callbackText = "✅ Удаление отменено"
	} else {
		// Неизвестный callback - отвечаем с ошибкой
		callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Неизвестная команда")
		bot.Request(callbackConfig)
		return
	}

	// Отвечаем на callback с текстом (покажет уведомление пользователю)
	callbackConfig := tgbotapi.NewCallback(callbackID, callbackText)
	bot.Request(callbackConfig)

	// Обновляем сообщение
	edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
	// Убираем клавиатуру после действия
	edit.ReplyMarkup = nil
	if _, err := bot.Send(edit); err != nil {
		log.Printf("Ошибка обновления сообщения: %v", err)
	}
}

// handleLanguageChange обрабатывает изменение языка интерфейса
func handleLanguageChange(bot *tgbotapi.BotAPI, callbackID string, chatID int64, messageID int, data string) {
	var langCode string
	var langText string

	var callbackText string

	switch data {
	case "lang_ru":
		langCode = "ru"
		langText = "🇷🇺 Язык изменён на Русский"
		callbackText = "✅ Язык изменён на Русский"
	case "lang_en":
		langCode = "en"
		langText = "🇬🇧 Language changed to English"
		callbackText = "✅ Language changed to English"
	case "lang_zh":
		langCode = "zh"
		langText = "🇨🇳 语言已更改为中文"
		callbackText = "✅ 语言已更改"
	default:
		// Неизвестный callback - отвечаем с ошибкой
		callbackConfig := tgbotapi.NewCallback(callbackID, "❌ Неизвестная команда")
		bot.Request(callbackConfig)
		return
	}

	// Сохраняем выбранный язык в памяти
	languagesMu.Lock()
	userLanguages[chatID] = langCode
	languagesMu.Unlock()

	// Отвечаем на callback с текстом (покажет уведомление пользователю)
	callbackConfig := tgbotapi.NewCallback(callbackID, callbackText)
	bot.Request(callbackConfig)

	// Обновляем сообщение с новым языком и клавиатурой
	editText := "🌐 Выбор языка:\n\n" + langText + "\n\nВыберите язык:"
	edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
	kb := keyboard.NewLanguageInlineKeyboard(langCode)
	kb = keyboard.AddBackButton(kb)
	edit.ReplyMarkup = &kb
	if _, err := bot.Send(edit); err != nil {
		log.Printf("Ошибка обновления сообщения: %v", err)
	}
}

// getLanguage возвращает текущий язык пользователя
// Если язык не сохранён, возвращает "ru" (по умолчанию русский)
func getLanguage(chatID int64) string {
	languagesMu.RLock()
	defer languagesMu.RUnlock()

	lang, exists := userLanguages[chatID]
	if !exists {
		// Если язык не сохранён, возвращаем значение по умолчанию (русский)
		return "ru"
	}
	return lang
}

// getNotificationState возвращает текущее состояние уведомлений для пользователя
// Если состояние не сохранено, возвращает true (по умолчанию включено)
func getNotificationState(chatID int64) bool {
	notificationsMu.RLock()
	defer notificationsMu.RUnlock()

	state, exists := userNotifications[chatID]
	if !exists {
		// Если состояние не сохранено, возвращаем значение по умолчанию (включено)
		return true
	}
	return state
}
