let router = null

export function setRouter(r) {
  router = r
}

export async function api(path, opts = {}) {
  const token = localStorage.getItem('token')
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', Authorization: token },
    ...opts,
  })
  if (res.status === 401) {
    localStorage.removeItem('token')
    if (router) router.push('/login')
    throw new Error('Session expired')
  }
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}
