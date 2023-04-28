package service

func getKeyboard(str string) string {
	mKeyboard := map[string]string{
		"start": `{
			"one_time": false,
			"buttons": [
			 [
			  {
			   "action": {
				"type": "text",
				"label": "ИАГ"
				},
				"color": "primary"
				}	
			]
		]
		}`,
		"institute": `{
			"one_time": false,
			"buttons": [
			 [
			  {
			   "action": {
				"type": "text",
				"label": "ИАГ"
			   },
			   "color": "primary"
			  },
			  {
			   "action": {
				"type": "text",
				"label": "ИПГС"
			   },
			   "color": "primary"
			  }
			 ],
			 [
			  {
			   "action": {
				"type": "text",
				"label": "ИГЭС"
			   },
			   "color": "primary"
			  },
			  {
			   "action": {
				"type": "text",
				"label": "ИИЭСМ"
			   },
			   "color": "primary"
			  }
			 ],
			 [
			  {
			   "action": {
				"type": "text",
				"label": "ИЭУКСН"
			   },
			   "color": "primary"
			  },
			  {
			   "action": {
				"type": "text",
				"label": "ИЦТМС"
			   },
			   "color": "primary"
			  }
			 ],
			 [
			  {
			   "action": {
				"type": "text",
				"label": "Вернуться"
			   },
			   "color": "secondary"
			  }
			 ]
			]
		   }`,
		"course": `{
		"one_time": false,
		"buttons": [
		 [
		  {
		   "action": {
			"type": "text",
			"label": "1 курс"
		   },
		   "color": "primary"
		  },
		  {
		   "action": {
			"type": "text",
			"label": "2 курс"
		   },
		   "color": "primary"
		  }
		 ],
		 [
		  {
		   "action": {
			"type": "text",
			"label": "3 курс"
		   },
		   "color": "primary"
		  },
		  {
		   "action": {
			"type": "text",
			"label": "4 курс"
		   },
		   "color": "primary"
		  }
		 ],
		 [
		  {
		   "action": {
			"type": "text",
			"label": "Вернуться"
		   },
		   "color": "secondary"
		  }
		 ]
		]
	   }`,
		"week": `{
			"one_time": false,
			"buttons": [
			 [
			  {
			   "action": {
				"type": "text",
				"label": "Нечетная неделя"
			   },
			   "color": "negative"
			  },
			  {
			   "action": {
				"type": "text",
				"label": "Четная неделя"
			   },
			   "color": "positive"
			  }
			 ]
			]
		   }`,
		"oddweek": `{
	 "one_time": false,
	 "buttons": [
	  [
	   {
		"action": {
		 "type": "text",
		 "label": "Понедельник"
		},
		"color": "negative"
	   },
	   {
		"action": {
		 "type": "text",
		 "label": "Вторник"
		},
		"color": "negative"
	   }
	  ],
	  [
	   {
		"action": {
		 "type": "text",
		 "label": "Среда"
		},
		"color": "negative"
	   },
	   {
		"action": {
		 "type": "text",
		 "label": "Четверг"
		},
		"color": "negative"
	   }
	  ],
	  [
	   {
		"action": {
		 "type": "text",
		 "label": "Пятница"
		},
		"color": "negative"
	   },
	   {
		"action": {
		 "type": "text",
		 "label": "Вернуться"
		},
		"color": "secondary"
	   }
	  ]
	 ]
	}`,
		"evenweek": `{
	 "one_time": false,
	 "buttons": [
	  [
	   {
		"action": {
		 "type": "text",
		 "label": "Понедельник"
		},
		"color": "positive"
	   },
	   {
		"action": {
		 "type": "text",
		 "label": "Вторник"
		},
		"color": "positive"
	   }
	  ],
	  [
	   {
		"action": {
		 "type": "text",
		 "label": "Среда"
		},
		"color": "positive"
	   },
	   {
		"action": {
		 "type": "text",
		 "label": "Четверг"
		},
		"color": "positive"
	   }
	  ],
	  [
	   {
		"action": {
		 "type": "text",
		 "label": "Пятница"
		},
		"color": "positive"
	   },
	   {
		"action": {
		 "type": "text",
		 "label": "Вернуться"
		},
		"color": "secondary"
	   }
	  ]
	 ]
	}`,
	}

	return mKeyboard[str]
}
