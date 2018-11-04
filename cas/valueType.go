package cas

// Cas 服务器返回
type CasReqReturn struct {
	ServiceResponse serviceResponse	`json:"serviceResponse"`
}
type serviceResponse struct{
	AuthenticationFailure AuthenticationFailure	`json:"authenticationFailure"`
	AuthenticationSuccess AuthenticationSuccess	`json:"authenticationSuccess"`
}

type AuthenticationFailure struct {
	Code string `json:"code"`
	Description string `json:"description"`
}

type AuthenticationSuccess struct {
	User string	`json:"user"`
	Attributes Attributes `json:"attributes"`
	Timeout int64	`json:"timeout"`
	Ticket string `json:"ticket"`
	Tgt string `json:"tgt"`
	//Service string `json:"service"`
}

type Attributes struct {
	Email string `json:"email"`
	Emid int64 `json:"emid"`
	EntCode string `json:"entCode"`
	EntId int64 `json:"entId"`
	Id int64 `json:"id"`
	Phone string `json:"phone"`
	Tgt	string `json:"tgt"`
	UserName string `json:"userName"`
	Name string `json:"name"`
	Type string `json:"type"`
	Sex string `json:"sex"`

}

