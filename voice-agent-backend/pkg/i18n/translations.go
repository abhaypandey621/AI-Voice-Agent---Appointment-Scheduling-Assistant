package i18n

// Language represents supported languages
type Language string

const (
	LanguageEnglish  Language = "en"
	LanguageSpanish  Language = "es"
	LanguageFrench   Language = "fr"
	LanguageGerman   Language = "de"
	LanguageHindi    Language = "hi"
	LanguageJapanese Language = "ja"
	LanguageChinese  Language = "zh"
)

// Translation strings for different languages
type Translations struct {
	Greetings      map[Language]string
	Prompts        map[Language]map[string]string
	Confirmations  map[Language]map[string]string
	Errors         map[Language]map[string]string
	SystemMessages map[Language]map[string]string
}

// NewTranslations creates a new translations object
func NewTranslations() *Translations {
	return &Translations{
		Greetings: map[Language]string{
			LanguageEnglish:  "Hello! I'm your AI appointment assistant. How can I help you today?",
			LanguageSpanish:  "¡Hola! Soy tu asistente de citas de IA. ¿Cómo puedo ayudarte hoy?",
			LanguageFrench:   "Bonjour! Je suis votre assistant de rendez-vous IA. Comment puis-je vous aider aujourd'hui?",
			LanguageGerman:   "Hallo! Ich bin dein KI-Termin-Assistent. Wie kann ich dir heute helfen?",
			LanguageHindi:    "नमस्ते! मैं आपका एआई अपॉइंटमेंट असिस्टेंट हूँ। मैं आपकी कैसे मदद कर सकता हूँ?",
			LanguageJapanese: "こんにちは！私はあなたのAIアポイントメントアシスタントです。今日はどうお手伝いしましょうか？",
			LanguageChinese:  "你好！我是你的人工智能预约助手。我今天能帮你什么？",
		},
		Prompts: map[Language]map[string]string{
			LanguageEnglish: {
				"ask_phone":               "Please provide your phone number so I can identify you.",
				"ask_appointment_time":    "What time would you like to book your appointment?",
				"ask_appointment_date":    "What date would you prefer for your appointment?",
				"ask_appointment_purpose": "What is the purpose of your appointment?",
				"confirm_booking":         "I'll book an appointment for you on %s at %s. Is that correct?",
				"ask_name":                "May I have your name please?",
				"ask_modification":        "What would you like to modify about your appointment?",
			},
			LanguageSpanish: {
				"ask_phone":               "Por favor, proporcione su número de teléfono para identificarle.",
				"ask_appointment_time":    "¿A qué hora le gustaría reservar su cita?",
				"ask_appointment_date":    "¿Qué fecha prefiere para su cita?",
				"ask_appointment_purpose": "¿Cuál es el propósito de su cita?",
				"confirm_booking":         "Reservaré una cita para usted el %s a las %s. ¿Es correcto?",
				"ask_name":                "¿Podría darme su nombre, por favor?",
				"ask_modification":        "¿Qué le gustaría modificar de su cita?",
			},
			LanguageFrench: {
				"ask_phone":               "Veuillez fournir votre numéro de téléphone pour que je vous identifie.",
				"ask_appointment_time":    "À quelle heure souhaiteriez-vous réserver votre rendez-vous?",
				"ask_appointment_date":    "Quelle date préférez-vous pour votre rendez-vous?",
				"ask_appointment_purpose": "Quel est l'objet de votre rendez-vous?",
				"confirm_booking":         "Je vais vous réserver un rendez-vous le %s à %s. Est-ce correct?",
				"ask_name":                "Puis-je avoir votre nom, s'il vous plaît?",
				"ask_modification":        "Que souhaitez-vous modifier dans votre rendez-vous?",
			},
			LanguageHindi: {
				"ask_phone":               "कृपया आपने पहचान करने के लिए अपना फोन नंबर प्रदान करें।",
				"ask_appointment_time":    "आप अपनी अपॉइंटमेंट के लिए किस समय बुक करना चाहते हैं?",
				"ask_appointment_date":    "आप अपनी अपॉइंटमेंट के लिए कौन सी तारीख पसंद करते हैं?",
				"ask_appointment_purpose": "आपकी अपॉइंटमेंट का उद्देश्य क्या है?",
				"confirm_booking":         "मैं आपकी %s को %s पर अपॉइंटमेंट बुक करूंगा। क्या यह सही है?",
				"ask_name":                "क्या आप मुझे अपना नाम बता सकते हैं?",
				"ask_modification":        "आप अपनी अपॉइंटमेंट में क्या संशोधन करना चाहते हैं?",
			},
		},
		Confirmations: map[Language]map[string]string{
			LanguageEnglish: {
				"booking_confirmed":      "Your appointment has been confirmed! You'll receive a reminder 24 hours before.",
				"cancellation_confirmed": "Your appointment has been successfully cancelled.",
				"modification_confirmed": "Your appointment has been updated successfully.",
			},
			LanguageSpanish: {
				"booking_confirmed":      "¡Su cita ha sido confirmada! Recibirá un recordatorio 24 horas antes.",
				"cancellation_confirmed": "Su cita ha sido cancelada exitosamente.",
				"modification_confirmed": "Su cita ha sido actualizada exitosamente.",
			},
			LanguageFrench: {
				"booking_confirmed":      "Votre rendez-vous a été confirmé! Vous recevrez un rappel 24 heures avant.",
				"cancellation_confirmed": "Votre rendez-vous a été annulé avec succès.",
				"modification_confirmed": "Votre rendez-vous a été mis à jour avec succès.",
			},
			LanguageHindi: {
				"booking_confirmed":      "आपकी अपॉइंटमेंट की पुष्टि हो गई है! आपको 24 घंटे पहले एक अनुस्मारक मिलेगा।",
				"cancellation_confirmed": "आपकी अपॉइंटमेंट सफलतापूर्वक रद्द कर दी गई है।",
				"modification_confirmed": "आपकी अपॉइंटमेंट सफलतापूर्वक अपडेट कर दी गई है।",
			},
		},
		Errors: map[Language]map[string]string{
			LanguageEnglish: {
				"invalid_phone":         "The phone number you provided is invalid. Please try again.",
				"slot_unavailable":      "The selected time slot is not available. Please choose another time.",
				"user_not_found":        "User not found. Please provide a valid phone number.",
				"appointment_not_found": "Appointment not found. Please check the details.",
				"double_booking":        "This time slot is already booked. Please select another time.",
			},
			LanguageSpanish: {
				"invalid_phone":         "El número de teléfono que proporcionó no es válido. Por favor, intente de nuevo.",
				"slot_unavailable":      "La hora seleccionada no está disponible. Por favor, elija otro tiempo.",
				"user_not_found":        "Usuario no encontrado. Por favor proporcione un número de teléfono válido.",
				"appointment_not_found": "Cita no encontrada. Por favor verifique los detalles.",
				"double_booking":        "Esta hora ya está reservada. Por favor seleccione otro tiempo.",
			},
			LanguageFrench: {
				"invalid_phone":         "Le numéro de téléphone que vous avez fourni est invalide. Veuillez réessayer.",
				"slot_unavailable":      "Le créneau horaire sélectionné n'est pas disponible. Veuillez choisir un autre créneau.",
				"user_not_found":        "Utilisateur non trouvé. Veuillez fournir un numéro de téléphone valide.",
				"appointment_not_found": "Rendez-vous non trouvé. Veuillez vérifier les détails.",
				"double_booking":        "Ce créneau horaire est déjà réservé. Veuillez sélectionner un autre créneau.",
			},
			LanguageHindi: {
				"invalid_phone":         "आपके द्वारा प्रदान किया गया फोन नंबर अमान्य है। कृपया फिर से प्रयास करें।",
				"slot_unavailable":      "चयनित समय स्लॉट उपलब्ध नहीं है। कृपया दूसरा समय चुनें।",
				"user_not_found":        "उपयोगकर्ता नहीं मिला। कृपया एक वैध फोन नंबर प्रदान करें।",
				"appointment_not_found": "अपॉइंटमेंट नहीं मिली। कृपया विवरण जांचें।",
				"double_booking":        "यह समय स्लॉट पहले से बुक है। कृपया दूसरा समय चुनें।",
			},
		},
		SystemMessages: map[Language]map[string]string{
			LanguageEnglish: {
				"call_started":    "Call started. Listening...",
				"call_ended":      "Call ended. Thank you for using our service.",
				"processing":      "Processing your request...",
				"available_slots": "Here are the available time slots for %s:",
			},
			LanguageSpanish: {
				"call_started":    "Llamada iniciada. Escuchando...",
				"call_ended":      "Llamada finalizada. Gracias por usar nuestro servicio.",
				"processing":      "Procesando su solicitud...",
				"available_slots": "Aquí están las franjas horarias disponibles para %s:",
			},
			LanguageFrench: {
				"call_started":    "Appel commencé. Écoute...",
				"call_ended":      "Appel terminé. Merci d'avoir utilisé notre service.",
				"processing":      "Traitement de votre demande...",
				"available_slots": "Voici les créneaux horaires disponibles pour %s:",
			},
			LanguageHindi: {
				"call_started":    "कॉल शुरू हुई। सुन रहे हैं...",
				"call_ended":      "कॉल समाप्त हुई। हमारी सेवा का उपयोग करने के लिए धन्यवाद।",
				"processing":      "आपके अनुरोध को संसाधित कर रहे हैं...",
				"available_slots": "%s के लिए यहां उपलब्ध समय स्लॉट दिए गए हैं:",
			},
		},
	}
}

// GetTranslation retrieves a translation for a specific key and language
func (t *Translations) GetTranslation(language Language, category string, key string) string {
	switch category {
	case "greeting":
		if msg, ok := t.Greetings[language]; ok {
			return msg
		}
		return t.Greetings[LanguageEnglish]
	case "prompt":
		if categoryMap, ok := t.Prompts[language]; ok {
			if msg, ok := categoryMap[key]; ok {
				return msg
			}
		}
		return t.Prompts[LanguageEnglish][key]
	case "confirmation":
		if categoryMap, ok := t.Confirmations[language]; ok {
			if msg, ok := categoryMap[key]; ok {
				return msg
			}
		}
		return t.Confirmations[LanguageEnglish][key]
	case "error":
		if categoryMap, ok := t.Errors[language]; ok {
			if msg, ok := categoryMap[key]; ok {
				return msg
			}
		}
		return t.Errors[LanguageEnglish][key]
	case "system":
		if categoryMap, ok := t.SystemMessages[language]; ok {
			if msg, ok := categoryMap[key]; ok {
				return msg
			}
		}
		return t.SystemMessages[LanguageEnglish][key]
	}
	return ""
}

// GetSupportedLanguages returns a list of all supported languages
func GetSupportedLanguages() []Language {
	return []Language{
		LanguageEnglish,
		LanguageSpanish,
		LanguageFrench,
		LanguageGerman,
		LanguageHindi,
		LanguageJapanese,
		LanguageChinese,
	}
}

// LanguageToCode converts language to language code (e.g., for speech synthesis)
func LanguageToCode(lang Language) string {
	switch lang {
	case LanguageEnglish:
		return "en-US"
	case LanguageSpanish:
		return "es-ES"
	case LanguageFrench:
		return "fr-FR"
	case LanguageGerman:
		return "de-DE"
	case LanguageHindi:
		return "hi-IN"
	case LanguageJapanese:
		return "ja-JP"
	case LanguageChinese:
		return "zh-CN"
	default:
		return "en-US"
	}
}

// DetectLanguageFromCode converts language code to Language (e.g., from speech recognition)
func DetectLanguageFromCode(code string) Language {
	switch code {
	case "es", "es-ES":
		return LanguageSpanish
	case "fr", "fr-FR":
		return LanguageFrench
	case "de", "de-DE":
		return LanguageGerman
	case "hi", "hi-IN":
		return LanguageHindi
	case "ja", "ja-JP":
		return LanguageJapanese
	case "zh", "zh-CN":
		return LanguageChinese
	default:
		return LanguageEnglish
	}
}
