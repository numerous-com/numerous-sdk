import streamlit as st

st.title("Counter Example")
count = 0

increment = st.button("Increment")
if increment:
    count += 1

st.write("Count = ", count)
