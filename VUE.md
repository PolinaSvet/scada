# SCADA System Architecture Documentation

## 🏗️ Общая архитектура системы

```
WebSocket Server → Pinia Store → Компоненты → DOM Анимация
```

## 📊 Поток данных

### 1. 📨 Входные данные (WebSocket)

**Файл:** `stores/websocket.js`

**Что происходит:**
- Подключение к WebSocket серверу
- Прием данных в реальном времени
- Первичная обработка сообщений

```javascript
// Входящие данные с сервера
{
  "type": "vuejsway",
  "data": {
    "id": "sensor_0",
    "info": { ... },
    "objVue": {
      "state": 7,
      "stateColor": "#C0C0C0",
      "stateTxt": "НЕОПРЕДЕЛЕНО", 
      "inputValue": "true °C",
      "ack": true,
      "mask": false
    },
    "rawData": { ... }
  }
}
```

**Обработка в `processMessage()`:**
```javascript
case 'vuejsway':
  objectsStore.updateObject(messageData) // → Передаем в хранилище объектов
  break;
```

### 2. 💾 Хранилище данных (Pinia)

**Файл:** `stores/objects.js`

**Что происходит:**
- Централизованное хранение состояния всех объектов
- Обновление данных объектов
- Управление активными элементами

```javascript
// Обновление объекта
updateObject(objData) {
  if (objData && objData.id) {
    this.objects[objData.id] = {
      ...this.objects[objData.id],
      ...objData
    }
  }
}
```

### 3. 🎯 Динамические компоненты

#### Файл: `components/ObjSensor.vue`

**Структура компонента:**
```vue
<template>
  <div class="obj-sensor" :style="sensorStyle" @click="handleClick">
    <svg>
      <!-- Анимированные элементы -->
      <rect :fill="stateColor" :stroke="borderColor" :class="{ blinking: shouldBlink }"/>
      <text>{{ stateTxt }}</text>
      <text>{{ inputValue }}</text>
    </svg>
  </div>
</template>
```

**Вычисляемые свойства (computed):**
```javascript
// Преобразование данных в визуальные свойства
stateColor() {
  return this.objVue.stateColor || '#C0C0C0'
},
borderColor() {
  return this.objVue.mask ? '#FF0000' : '#000000'
},
shouldBlink() {
  return this.objVue.ack === false  // Мигание при отсутствии квитирования
}
```

## 🎨 Создание новых мнемосхем

### Шаг 1: Добавление SVG схемы

**Файл:** `assets/schemas/новая_схема.svg`

**Требования к SVG:**
```xml
<svg width="800" height="600" viewBox="0 0 800 600">
  <!-- Статический фон -->
  <rect width="100%" height="100%" fill="#1e3a5f"/>
  
  <!-- Динамические объекты (якоря) -->
  <rect x="100" y="100" width="120" height="80" 
        data-type="obj-sensor-anchor" 
        data-tag="NEW_SENSOR_1" 
        data-id="new_sensor_1"/>
        
  <circle cx="400" cy="300" r="40"
          data-type="obj-sensor-anchor"
          data-tag="NEW_SENSOR_2" 
          data-id="new_sensor_2"/>
</svg>
```

**Обязательные атрибуты якорей:**
- `data-type="obj-sensor-anchor"` - тип элемента
- `data-tag="TAG_NAME"` - технологический тег
- `data-id="object_id"` - уникальный ID объекта

### Шаг 2: Регистрация схемы в хранилище

**Файл:** `stores/schemas.js`

```javascript
const schemas = ref([
  {
    id: 'new_schema',
    name: 'Новая мнемосхема',
    description: 'Описание новой схемы',
    svgFile: 'новая_схема.svg'
  },
  // ... существующие схемы
])
```

## 🔧 Создание новых типов объектов

### Вариант 1: Использование существующего ObjSensor

**Для простых объектов** достаточно добавить якоря в SVG. Компонент `ObjSensor.vue` автоматически обработает:
- Цвет фона (`stateColor`)
- Цвет рамки (`mask`)
- Мигание (`ack`)
- Текст состояния (`stateTxt`)
- Значение (`inputValue`)

### Вариант 2: Создание нового компонента

**Файл:** `components/ObjPump.vue` (пример насоса)

```vue
<template>
  <div class="obj-pump" :style="pumpStyle" @click="handleClick">
    <svg>
      <!-- Специфичная для насоса SVG графика -->
      <circle :fill="statusColor" :class="{ rotating: isRunning }"/>
      <text>{{ flowRate }} m³/h</text>
    </svg>
  </div>
</template>

<script>
export default {
  props: ['id', 'tag', 'x', 'y'],
  
  computed: {
    pumpData() {
      return this.$objectsStore.getObject(this.id)
    },
    
    statusColor() {
      const status = this.pumpData?.objVue?.state
      return status === 'running' ? '#00FF00' : '#FF0000'
    },
    
    isRunning() {
      return this.pumpData?.objVue?.state === 'running'
    },
    
    flowRate() {
      return this.pumpData?.objVue?.inputValue || '0'
    }
  }
}
</script>

<style>
.rotating {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
```

## 🔄 Полный цикл обработки данных

### Пример для датчика температуры:

1. **Сервер отправляет данные:**
```json
{
  "id": "sensor_0",
  "objVue": {
    "stateColor": "#FF0000",
    "stateTxt": "АВАРИЯ",
    "inputValue": "95.5 °C",
    "ack": false,
    "mask": true
  }
}
```

2. **WebSocket получает и передает в хранилище**
3. **Pinia обновляет объект `sensor_0`**
4. **ObjSensor автоматически перерисовывается:**
   - Фон становится красным (`#FF0000`)
   - Рамка красная (`mask: true`)
   - Начинает мигать (`ack: false`)
   - Отображается "АВАРИЯ" и "95.5 °C"

## 🎮 Управление объектами

### Окно управления (`ObjSensorCtrl.vue`)

**Активация:** Клик на объекте с `ctrlEnable: true`

**Данные для отображения:**
- Основная информация из `info`
- Сырые данные из `rawData`
- Текущее состояние из `objVue`

## 📱 Навигация между схемами

### Файл: `components/NavigationBar.vue`

**Механика переключения:**
1. Пользователь кликает на кнопку схемы
2. `schemasStore.setCurrentSchema(schemaId)`
3. `SchemaViewer` загружает соответствующую SVG
4. `MnemoSchema` парсит якоря и создает компоненты

## 🛠️ Отладка и разработка

### Ключевые точки для отладки:

1. **WebSocket соединение:** Консоль браузера → Сеть → WS
2. **Данные объектов:** Vue DevTools → Pinia → objects
3. **Позиционирование:** Временные рамки вокруг компонентов
4. **Анимации:** Vue DevTools → Компоненты → ObjSensor

### Полезные команды в консоли:
```javascript
// Проверить данные объекта
$objectsStore.objects['sensor_0']

// Проверить текущую схему
$schemasStore.currentSchemaId

// Принудительно обновить схему
$schemasStore.loadSchemaSvg('schema1')
```

## 📈 Расширение системы

### Добавление нового типа данных:

1. **Создать компонент** в `components/ObjNewType.vue`
2. **Зарегистрировать в MnemoSchema.vue** (если отличается от sensor)
3. **Добавить обработку в SVG** через новый `data-type`
4. **Обновить парсинг** в `extractCoordsFromSvgString()`

### Добавление сложной анимации:

1. **Определить триггеры** в данных (`objVue.newProperty`)
2. **Добавить computed свойства** в компонент
3. **Создать CSS анимации** или использовать JavaScript
4. **Протестировать производительность** на большом количестве объектов

Эта архитектура обеспечивает масштабируемость и простоту добавления новых функциональностей while maintaining real-time performance.