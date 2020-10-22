# gotest
## API
#### `POST /images`
Загружает изображение из `image` поля формы в галерею. Формат ответа:
```
{
  'id': 1, 
  'parent': None,
  'basename': 'c6jn1fitxyuu.jpeg',
  'uri': '/static/c6jn1fitxyuu.jpeg',
  'width': 1920, 
  'height': 1080
}
```

#### `POST /images/:image_id/preview`
Генерирует preview нужного размера и детальности для `image_id`. Параметры:
```
width: int | 28 <= width <= parent.width - Превалирует над height в аспекте пропорций;
height: int | 28 <= height <= parent.height;
resampling: str | str ∈ (lanczos, catmull, mitnet, linear, box, nearest)
```
Отвечает в одном с `POST /images` формате.

#### Про статику
`POST /images` и `POST /images/:image_id/preview` дополнительно генерируют .gz версии изображений.
Предполагается, что статикой будет заведовать front-сервер. Приложение ищет БД и сохраняет файлы в `/opt/persist` контейнера.

#### `GET /images`
Возвращает список всех картинок в галерее.

#### `DELETE /images/:image_id`
Удаляет картинку с ключом `image_id`. Удаление картинки с `parent: None` влечет за собой удаление всех ее preview.
```
