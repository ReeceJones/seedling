
export async function login(email: string, password: string): Promise<string> {
  const r = await fetch('http://localhost:8081/v1/users/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
  })
  if (!r.ok) {
    throw new Error('Invalid username or password')
  }

  const json_body = await r.json()
  const token = json_body.token

  localStorage.setItem('token', token)

  return token
}
