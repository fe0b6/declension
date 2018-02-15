package declension

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	rules   map[string]map[string][]ruleData
	genders gender
	cases   map[string]int
)

func init() {
	cases = map[string]int{
		"ИП": -1,
		"РП": 0,
		"ДП": 1,
		"ВП": 2,
		"ТП": 3,
		"ПП": 4,
	}
}

// Init - инициализация словарей
func Init(ruleFile, genderFile string) {

	d1, err := ioutil.ReadFile(ruleFile)
	if err != nil {
		log.Println("[error]", err)
		return
	}

	err = json.Unmarshal(d1, &rules)
	if err != nil {
		log.Println("[error]", err)
		return
	}

	d2, err := ioutil.ReadFile(genderFile)
	if err != nil {
		log.Println("[error]", err)
		return
	}

	err = json.Unmarshal(d2, &genders)
	if err != nil {
		log.Println("[error]", err)
		return
	}
}

// Fio - Возвращаем склоненное ФИО
func Fio(name, c, gen string) (newName string, err error) {
	names := strings.Split(name, " ")
	// Определяем пол
	if gen == "" {
		gen = GetGender(names[0], "lastname")
	}
	if gen == "" {
		gen = GetGender(names[1], "firstname")
	}
	if gen == "" {
		gen = GetGender(names[2], "middlename")
	}
	if gen == "" {
		err = errors.New("Не удалось определить пол")
		return
	}

	fname, err := inflect(names[1], "firstname", cases[c])
	if err != nil {
		return
	}

	lname, err := inflect(names[0], "lastname", cases[c])
	if err != nil {
		return
	}

	mname, err := inflect(names[2], "middlename", cases[c])
	if err != nil {
		return
	}

	newName = strings.Join([]string{lname, fname, mname}, " ")

	return
}

// Word - Возвращаем склоненное слово
func Word(word, c, t, gen string) (newWord string, err error) {
	// Определяем пол
	if gen == "" {
		gen = GetGender(word, t)
	}

	newWord, err = inflect(word, t, cases[c])
	if err != nil {
		return
	}

	return
}

// Words - Возвращаем склоненные слова
func Words(word, c, t, gen string) (newWord string, err error) {
	arr := []string{}

	for _, w := range strings.Split(word, " ") {
		// Определяем пол
		if gen == "" {
			gen = GetGender(w, t)
		}

		var nw string
		nw, err = inflect(w, t, cases[c])
		if err != nil {
			return
		}

		arr = append(arr, nw)
	}

	newWord = strings.Join(arr, " ")
	return
}

// Применяем правило
func applyRule(name, mod string) (result string, err error) {
	arr := []string{}
	l := utf8.RuneCountInString(name) - strings.Count(mod, "-")
	for i, v := range strings.Split(name, "") {
		if i >= l {
			break
		}

		arr = append(arr, v)
	}

	result = strings.Join(arr, "")
	result += strings.Replace(mod, "-", "", -1)
	return
}

// Трансформируем слово
func inflect(name, t string, c int) (newName string, err error) {
	newName, err = checkException(name, t, c)
	if err != nil || newName != "" {
		return
	}

	newName, err = findInRules(name, t, c)
	if err != nil {
		return
	}

	if newName == "" {
		err = errors.New("Склонение не удалось")
		log.Println(err, name)
	}

	return
}

// Проверяем исключения
func checkException(name, t string, c int) (newName string, err error) {
	ex := rules[t]["exceptions"]
	if len(ex) == 0 {
		return
	}

	lowerName := strings.ToLower(name)
	for _, r := range ex {
		if !checkGender(r.Gender, name, t) {
			continue
		}

		for _, w := range r.Test {
			var matched bool
			matched, err = regexp.MatchString(regexp.QuoteMeta(strings.ToLower(w))+"$", lowerName)
			if err != nil {
				log.Println("[error]", err)
				continue
			}

			if !matched {
				continue
			}

			if r.Mods[c] == "." {
				newName = name
				return
			}

			return applyRule(name, r.Mods[c])
		}
	}

	return
}

// Проверяем правила
func findInRules(name, t string, c int) (newName string, err error) {
	ex := rules[t]["suffixes"]
	if len(ex) == 0 {
		return
	}

	lowerName := strings.ToLower(name)
	for _, r := range ex {
		if !checkGender(r.Gender, name, t) {
			continue
		}

		for _, w := range r.Test {
			var matched bool
			matched, err = regexp.MatchString(regexp.QuoteMeta(strings.ToLower(w))+"$", lowerName)
			if err != nil {
				log.Println("[error]", err)
				continue
			}

			if !matched {
				continue
			}

			if r.Mods[c] == "." {
				newName = name
				return
			}

			return applyRule(name, r.Mods[c])
		}
	}

	return
}

// Проверяем подходит ли пол
func checkGender(gen, name, t string) (ok bool) {
	if gen == GetGender(name, t) {
		ok = true
	}

	return
}

// GetGender - Получаем пол. t = какая часть ФИО проверяется
func GetGender(name, t string) (gen string) {
	for k, arr := range genders.Gender[t] {
		for _, en := range arr {
			match, err := regexp.MatchString(regexp.QuoteMeta(en)+"$", name)
			if err != nil {
				log.Println("[error]", err)
				continue
			}

			if match {
				gen = k
				return
			}
		}
	}

	return
}
