# gotest
## API
#### `POST /api/v1/images`
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

#### `POST /api/v1/images/:image_id/preview`
Генерирует preview нужного размера и детальности для `image_id`. Параметры:
```
width: int | 28 <= width <= parent.width - Превалирует над height в аспекте пропорций;
height: int | 28 <= height <= parent.height;
resampling: str | str ∈ (lanczos, catmull, mitnet, linear, box, nearest)
```
Отвечает в одном с `POST /api/v1/images` формате.

#### Про статику
`POST /api/v1/images` и `POST /api/v1/images/:image_id/preview` дополнительно генерируют .gz версии изображений.
Предполагается, что статикой будет заведовать front-сервер. Приложение ищет БД и сохраняет файлы в `/opt/persist` контейнера. Можно настроить через переменные окружения (см. docker-compose.yml).

#### `GET /api/v1/images`
Возвращает список всех картинок в галерее.

#### `DELETE /api/v1/images/:image_id`
Удаляет картинку с ключом `image_id`. Удаление картинки с `parent: None` влечет за собой удаление всех ее preview.
```
