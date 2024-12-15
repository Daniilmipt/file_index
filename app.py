import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
import streamlit as st
from sklearn.linear_model import LinearRegression

Singapore = pd.read_csv('data.csv')
Singapore['Unnamed: 0'] = pd.to_datetime(Singapore['Unnamed: 0'])

st.title("Домашка Шамигулов")
st.sidebar.title("Меню")

page = st.sidebar.radio("Выберите раздел:", ["Анализ данных", "Гипотеза потепления"])

if page == "Анализ данных":
    st.header("Анализ средних температур в Сингапуре")


    st.subheader("вид таблицы с данными про Сингапур")
    st.dataframe(Singapore)

    # Группировка по годам для получения средней температуры
    temp = Singapore.groupby('year')['AverageTemperature'].mean().reset_index()

    # Визуализация данных
    st.subheader("График средней температуры по годам")
    plt.figure(figsize=(10, 5))
    sns.lineplot(data=temp, x='year', y='AverageTemperature')
    plt.title('Средняя температура по годам')
    plt.xlabel('Год')
    plt.ylabel('Средняя температура')
    st.pyplot(plt)


    st.subheader("Boxplots температур в Сингапуре по месяцам")

    Singapore.loc[Singapore['month']=='1','month'] = 'January'
    Singapore.loc[Singapore['month']=='2','month'] = 'February'
    Singapore.loc[Singapore['month']=='3','month'] = 'March'
    Singapore.loc[Singapore['month']=='4','month'] = 'April'
    Singapore.loc[Singapore['month']=='5','month'] = 'May'
    Singapore.loc[Singapore['month']=='6','month'] = 'June'
    Singapore.loc[Singapore['month']=='7','month'] = 'July'
    Singapore.loc[Singapore['month']=='8','month'] = 'August'
    Singapore.loc[Singapore['month']=='9','month'] = 'September'
    Singapore.loc[Singapore['month']=='10','month'] = 'October'
    Singapore.loc[Singapore['month']=='11','month'] = 'November'
    Singapore.loc[Singapore['month']=='12','month'] = 'December'
    year_month = Singapore.groupby(by = ['year', 'month','season']).mean().reset_index()
    plt.figure(figsize=(16,12))

    sns.boxplot(x = 'month', y = 'AverageTemperature', data = year_month, palette = "RdBu", saturation = 1, width = 0.9, fliersize=4, linewidth=2)
    plt.title('Average Temperature by Months', fontsize = 25)
    plt.xlabel('Months', fontsize = 20)
    plt.ylabel('Temperature', fontsize = 20)
    st.pyplot(plt)



if page == "Гипотеза потепления":
    st.header("Линейная регрессия для проверки тренда")

    # Линейная регрессия
    temp = Singapore.groupby('year')['AverageTemperature'].mean().reset_index()

    X = temp['year'].values.reshape(-1, 1)
    y = temp['AverageTemperature'].values

    model = LinearRegression()
    model.fit(X, y)

    # Получение предсказаний и коэффициента наклона
    y_pred = model.predict(X)
    slope = model.coef_[0]

    # Визуализация линейной регрессии
    st.subheader("График линейной регрессии")
    plt.figure(figsize=(10, 5))
    sns.lineplot(data=temp, x='year', y='AverageTemperature', label='Средняя температура')
    plt.plot(temp['year'], y_pred, color='red', label='Линейная регрессия')
    plt.title('Линейная регрессия для проверки тренда')
    plt.xlabel('Год')
    plt.ylabel('Средняя температура')
    plt.legend()
    st.pyplot(plt)

    # Настройка параметров для графика
    st.subheader("Настройка параметров графика")

    year_start = st.slider("Выберите начальный год:", min_value=int(temp['year'].min()), max_value=int(temp['year'].max()), value=int(temp['year'].min()))
    year_end = st.slider("Выберите конечный год:", min_value=int(temp['year'].min()), max_value=int(temp['year'].max()), value=int(temp['year'].max()))

    # Фильтрация данных по выбранным годам
    filtered_data = temp[(temp['year'] >= year_start) & (temp['year'] <= year_end)]

    # Обновление графика на основе выбранных параметров
    plt.figure(figsize=(10, 5))
    sns.lineplot(data=filtered_data, x='year', y='AverageTemperature', label='Средняя температура')

    # Обучение модели на отфильтрованных данных
    X_filtered = filtered_data['year'].values.reshape(-1, 1)
    y_filtered = filtered_data['AverageTemperature'].values

    model_filtered = LinearRegression()
    model_filtered.fit(X_filtered, y_filtered)

    y_pred_filtered = model_filtered.predict(X_filtered)

    plt.plot(filtered_data['year'], y_pred_filtered, color='red', label='Линейная регрессия')

    plt.title('Линейная регрессия для выбранного диапазона годов')
    plt.xlabel('Год')
    plt.ylabel('Средняя температура')
    plt.legend()
    st.pyplot(plt)

    from scipy import stats
st.subheader("Код проверки гипотезы")

st.code(
    "slope = model.coef_[0]\n"
    "slope_t_statistic = slope / (np.std(y - y_pred) / np.sqrt(len(y)))\n"
    "p_value = stats.t.sf((slope_t_statistic), df=len(y)-2) * 2\n"
    "print(f'Коэффициент наклона: {slope}')\n"
    "print(f'p-value: {p_value}')\n"
    "if p_value > 0.05:\n"
    "    print('Не можем отвергнуть нулевую гипотезу: существует положительный тренд.')\n"
    "else:\n"
    "print('Отвергаем нулевую гипотезу: нет статистически значимого тренда.')\n"
)

from scipy import stats

slope = model.coef_[0]
slope_t_statistic = slope / (np.std(y - y_pred) / np.sqrt(len(y)))
p_value = stats.t.sf((slope_t_statistic), df=len(y)-2) * 2

print(f'Коэффициент наклона: {slope}')
print(f'p-value: {p_value}')

if p_value > 0.05:
    print("Не можем отвергнуть нулевую гипотезу: существует положительный тренд.")
else:
    print("Отвергаем нулевую гипотезу: нет статистически значимого тренда.")
