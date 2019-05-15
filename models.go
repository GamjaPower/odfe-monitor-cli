package main

// Search hello
type Search struct {
	Search struct {
		Indices []string               `json:"indices"`
		Query   map[string]interface{} `json:"query"`
	} `json:"search"`
}

type Trigger struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Severity  string    `json:"severity"`
	Condition Condition `json:"condition"`
	Actions   []Action  `json:"actions,omitempty"`
}

// Period hello
type Period struct {
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}

// Cron hello
type Cron struct {
	Expression string `json:"expression"`
	Timezone   string `json:"timezone"`
}

// Schedule world
type Schedule struct {
	Period *Period `json:"period,omitempty"`
	Cron   *Cron   `json:"cron,omitempty"`
}

//Action action model
type Action struct {
	Name            string `json:"name"`
	DestinationID   string `json:"destination_id,omitempty"`
	SubjectTemplate Script `json:"subject_template"`
	MessageTemplate Script `json:"message_template"`
}

type Script struct {
	Source string `json:"source"`
	Lang   string `json:"lang"`
}
type Condition struct {
	Script Script `json:"script"`
}

// Monitor nice
type Monitor struct {
	id       string
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Enabled  bool      `json:"enabled"`
	Schedule Schedule  `json:"schedule"`
	Inputs   []Search  `json:"inputs"`
	Triggers []Trigger `json:"triggers"`
}

type Config struct {
	Destinations map[string]string
}
