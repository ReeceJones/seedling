export interface Service {
  name: string
  description: string
  key: string
  project_url: string
  is_installed: boolean
  is_available: boolean
  is_running: boolean
  status: string
  icon: string
}

export interface InstalledService {
  name: string
  description: string
  live_url: string
  service: Service
}


export async function getInstalledServices(): Promise<InstalledService[]> {
  const token = localStorage.getItem('token')
  const r = await fetch('http://localhost:8081/v1/services/installed', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token,
    },
  })
  if (!r.ok) {
    throw new Error(await r.text())
  }

  const json_body = await r.json()
  const services = json_body.data

  return services
}

export async function getServices(): Promise<Service[]> {
  const token = localStorage.getItem('token')
  const r = await fetch('http://localhost:8081/v1/services/info', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token,
    },
  })
  if (!r.ok) {
    throw new Error(await r.text())
  }

  const json_body = await r.json()
  const services = json_body.data

  return services
}

export async function installService(key: string): Promise<InstalledService> {
  const token = localStorage.getItem('token')
  const r = await fetch('http://localhost:8081/v1/services/install', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token,
    },
    body: JSON.stringify({ key }),
  })
  if (!r.ok) {
    throw new Error(await r.text())
  }

  const json_body = await r.json()
  const installed_service = json_body.data

  return installed_service
}
