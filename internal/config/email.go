package config

var (
	AdminEmails = []string{"aleshus2007@gmail.com", "alekseyelcha07@mail.ru"}

	TicketCreatedMessageForUser = "<html><body>" +
		"Уважаемый (ая) %s, уведомляем Вас о том, " +
		"что Вы создали обращение на сайте ООО «Домофон-сервис» г. Одинцово.<br>" +
		"Следить за статусом обращения Вы можете на сайте <a href=\"https://domofon-servis-odi.ru\">ask.domofon-servis-odi.ru</a><br><br>" +
		"Это письмо было отправлено автоматически, просьба не отвечать на него." +
		"</body></html>"

	TicketCreatedMessageForAdmins = "На сайте было создано новое обращение: email: %s, name: %s, addresses: %s"

	TicketCreatedSubjectForUser   = "Обращение на сайте ООО «Домофон-сервис» г. Одинцово"
	TicketCreatedSubjectForAdmins = "Новое обращение на платформе"

	TicketAnsweredMessageForUser = "<html><body>" +
		"Уважаемый (ая) %s, на Ваше обращение был дан ответ.<br>" +
		"Ознакомиться с ним Вы можете на сайте <a href=\"https://domofon-servis-odi.ru\">ask.domofon-servis-odi.ru</a><br><br>" +
		"Это письмо было отправлено автоматически, просьба не отвечать на него." +
		"</body></html>"

	TicketAnsweredSubjectForUser = "Ответ на Ваше обращение"

	AuthCodeSentSubject = "Код для входа на сайт"
	AuthCodeSentMessage = "<html><body>" +
		"Код для входа на сайт: <b>%s</b>.<br><br>" +
		"Данное письмо было отправлено автоматически." +
		"</body></html>"

	//ExtraMessageCreatedMessageForUser   = ""
	//ExtraMessageCreatedMessageForAdmins = ""
)
