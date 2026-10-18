package config

var (
	AdminEmails = []string{"aleshus2007@gmail.com", "alekseyelcha07@mail.ru"}

	TicketCreatedMessageForUser = "Уважаемый (ая) %s, уведомляем Вас о том," +
		"что Вы создали обращение #%s на сайте ООО «Домофон-сервис» г. Одинцово. \n" +
		"Следить за статусом обращения Вы можете на сайте ask.domofon-servis-odi.ru\n\n" +
		"Это письмо было отправлено автоматически, просьба не отвечать на него."
	TicketCreatedMessageForAdmins = "На сайте было создано новое обращение."

	TicketCreatedSubjectForUser  = "Обращение на сайте ООО «Домофон-сервис» г. Одинцово"
	TicketCreatedSubjectForAdmin = "Новое обращение на платформе"

	TicketAnsweredMessageForUser = "Уважаемый (ая) %s, на Ваше обращение был дан ответ.\n" +
		"Ознакомиться с ним Вы можете на сайте ask.domofon-servis-odi.ru\n\n" +
		"Это письмо было отправлено автоматически, просьба не отвечать на него."

	//ExtraMessageCreatedMessageForUser   = ""
	//ExtraMessageCreatedMessageForAdmins = ""
)
