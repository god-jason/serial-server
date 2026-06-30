import { reactive } from 'vue'
import api from '../utils/api'

const session = reactive({
  isLoggedIn: false,
  checking: false
})

export async function checkLoginStatus() {
  session.checking = true
  try {
    await api.get('/system/info')
    session.isLoggedIn = true
  } catch (error) {
    session.isLoggedIn = false
  } finally {
    session.checking = false
  }
  return session.isLoggedIn
}

export function setLoggedIn(value) {
  session.isLoggedIn = value
}

export function getSession() {
  return session
}

export function isLoggedIn() {
  return session.isLoggedIn
}

export default session