export function courseName(code: string, title = code): string {
  return (
    (
      {
        ALG201: 'Алгоритмы',
        DBS202: 'Базы данных',
        WEB203: 'Веб-разработка',
        SWE204: 'Программная инженерия',
        NET205: 'Компьютерные сети',
        OSY206: 'Операционные системы',
        MLR207: 'Машинное обучение',
        MOB208: 'Мобильная разработка',
      } as Record<string, string>
    )[code] ?? title
  );
}

export function optionName(value: string): string {
  return (
    (
      {
        man: 'Мужчины',
        woman: 'Женщины',
        non_binary: 'Небинарные',
        other: 'Другое',
        everyone: 'Все',
        long_term_partner: 'Серьёзные отношения',
        long_term_open_to_short: 'Отношения, открыт(а) к общению',
        short_term_open_to_long_term: 'Общение, открыт(а) к отношениям',
        short_term: 'Ненадолго',
        new_friends: 'Дружба',
        still_figuring_it_out: 'Пока не знаю',
      } as Record<string, string>
    )[value] ?? value
  );
}
