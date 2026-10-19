package config

var (
	AdminEmails = []string{"aleshus2007@gmail.com", "alekseyelcha07@mail.ru"}

	TicketCreatedMessageForUser = "Уважаемый (ая) %s, уведомляем Вас о том," +
		"что Вы создали обращение на сайте ООО «Домофон-сервис» г. Одинцово.<br>" +
		"Следить за статусом обращения Вы можете на сайте ask.domofon-servis-odi.ru<br>" +
		"Это письмо было отправлено автоматически, просьба не отвечать на него."
	TicketCreatedMessageForAdmins = "На сайте было создано новое обращение: email: %s, name: %s, addresses: %s"

	TicketCreatedSubjectForUser  = "Обращение на сайте ООО «Домофон-сервис» г. Одинцово"
	TicketCreatedSubjectForAdmin = "Новое обращение на платформе"

	TicketAnsweredMessageForUser = "Уважаемый (ая) %s, на Ваше обращение был дан ответ.<br>" +
		"Ознакомиться с ним Вы можете на сайте ask.domofon-servis-odi.ru<br><br>" +
		"Это письмо было отправлено автоматически, просьба не отвечать на него."

	AuthCodeSentSubject = "Код для входа на сайт"
	AuthCodeSentMessage = "Код для входа на сайт: <b>%s</b> <br><br> Данное письмо было отправлено автоматически."

	//ExtraMessageCreatedMessageForUser   = ""
	//ExtraMessageCreatedMessageForAdmins = ""
)
