"use client"
import dynamic from "next/dynamic"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardFooter, CardDescription, CardTitle } from "@/components/ui/card"
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table"
import { Progress } from "@/components/ui/progress"
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card"
import { Pie, Cell, XAxis, YAxis, Line, ResponsiveContainer, Tooltip } from "recharts"
import {Alert, AlertTitle, AlertDescription} from "@/components/ui/alert"
import { ScrollArea } from "@/components/ui/scroll-area"
import Link from "next/link"
import Navbar from '@/components/navbar'

const LineChart = dynamic(
  () => import('recharts').then((mod) => mod.LineChart),
  { ssr: false }
)

const PieChart = dynamic(
  () => import('recharts').then((mod) => mod.PieChart),
  { ssr: false }
)

const services = [
  {
    name: "qBittorrent",
    status: "running",
    url: "http://localhost:8080"
  },
  {
    name: "Plex",
    status: "running",
    url: "http://localhost:32400"
  },
  {
    name: "Rclone",
    status: "running",
    url: "http://localhost:5572"
  },
  {
    name: "Radarr",
    status: "running",
    url: "http://localhost:7878"
  },
  {
    name: "Sonarr",
    status: "running",
    url: "http://localhost:8989"
  },
  {
    name: "Jackett",
    status: "running",
    url: "http://localhost:9117"
  },
  {
    name: "Overseerr",
    status: "running",
    url: "http://localhost:5055"
  }
]
const storage = [
  {
    name: "Free",
    value: 20
  },
  {
    name: "Used",
    value: 80
  }
]
const filesystems = [
  {
    name: "external_hdd",
    path: "/storage/0",
    total: 4600,
    used: 2000
  },
  {
    name: "google_drive",
    path: "/storage/1",
    total: 10,
    used: 8
  },
  {
    name: "nas",
    path: "/storage/2",
    total: 100000,
    used: 20000
  }
]
const network = [
  {
    time: "00:00",
    download: 10,
    upload: 1
  },
  {
    time: "01:00",
    download: 11,
    upload: 4
  },
  {
    time: "02:00",
    download: 10,
    upload: 3
  },
  {
    time: "03:00",
    download: 20,
    upload: 7
  },
  {
    time: "04:00",
    download: 30,
    upload: 4
  },
  {
    time: "05:00",
    download: 23,
    upload: 6
  },
  {
    time: "06:00",
    download: 25,
    upload: 6
  },
  {
    time: "07:00",
    download: 24,
    upload: 5
  }
]

function pieChartTooltip(d: any) {
  const {active, payload}: {active: boolean, payload: any} = d
  if (!active) {
    return <></>
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>{payload[0]['name']}</CardTitle>
      </CardHeader>
      <CardContent>
        <p>{payload[0]['value']} GB</p>
      </CardContent>
    </Card>
  )
}

function lineChartTooltip(d: any) {
  const {active, payload}: {active: boolean, payload: any} = d
  if (!active) {
    return <></>
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          Network Usage
        </CardTitle>
      </CardHeader>
      <CardContent>
        {
          payload.map((entry: any) => (
            <p key={entry.name}>{entry.name}: {entry.value} GB</p>
          ))
        }
      </CardContent>
    </Card>
  )
}

export default function Home() {
  return (
    <>
      <Navbar />
      <div className="w-full p-10">
        <div className="mx-auto w-[1200px]">
          <div className="m-4">
            <h1 className="text-4xl font-bold">Slimcat</h1>
          </div>
          <div className="grid grid-cols-5 w-[1200px]">
            <Card className="col-span-3 m-2">
              <CardHeader>
                <CardTitle>Status</CardTitle>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Service</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="text-right">URL</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {services.map((service) => (
                      <TableRow key={service.name}>
                        <TableCell><span className="font-bold">{service.name}</span></TableCell>
                        <TableCell><Badge>{service.status}</Badge></TableCell>
                        <TableCell className="text-right"><Link className="hover:underline transition" href={service.url}>{service.url}</Link></TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
            <Card className="col-span-2 m-2">
              <CardHeader>
                <CardTitle>Storage</CardTitle>
              </CardHeader>
              <CardContent>
                <p><span className="font-bold">Total:</span> {storage[0].value + storage[1].value} GB</p>
                <p><span className="font-bold">Used:</span> {storage[1].value} GB</p>
                <p><span className="font-bold">Free:</span> {storage[0].value} GB</p>
                <PieChart className="mx-auto" width={200} height={200}>
                  <Pie data={storage} dataKey="value" fill="black">
                    {
                      storage.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.name == "Used" ? "white" : "black"} />)
                    }
                  </Pie>
                  <Tooltip content={pieChartTooltip}/>
                </PieChart>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>File System</TableHead>
                      <TableHead>Path</TableHead>
                      <TableHead className="text-right">Used</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filesystems.map((filesystem) => (
                      <TableRow key={filesystem.name}>
                        <TableCell><span className="font-bold">{filesystem.name}</span></TableCell>
                        <TableCell>{filesystem.path}</TableCell>
                        <TableCell className="w-[150px]">
                          <HoverCard>
                            <HoverCardTrigger asChild>
                              <Progress value={filesystem.used * 100 / filesystem.total}/>
                            </HoverCardTrigger>
                            <HoverCardContent className="w-40">
                              <p><span className="font-bold">Total</span>: {filesystem.total} GB</p>
                              <p><span className="font-bold">Used</span>: {filesystem.used} GB</p>
                              <p><span className="font-bold">Free</span>: {filesystem.total - filesystem.used} GB</p>
                            </HoverCardContent>
                          </HoverCard>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
            <Card className="col-span-5 m-2">
              <CardHeader>
                <CardTitle>Network</CardTitle>
                <div><span className="align-middle">VPN</span> <Badge>active</Badge></div>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" className="p-0 m-0" height={400}>
                  <LineChart data={network}>
                    <XAxis dataKey="time" interval="preserveStartEnd"/>
                    <Tooltip content={lineChartTooltip}/>
                    <Line type="monotone" dataKey="download" stroke="#8884d8" strokeWidth={3} dot={false}/>
                    <Line type="monotone" dataKey="upload" stroke="#82ca9d" strokeWidth={3} dot={false}/>
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
            <Card className="col-span-4 m-2">
              <CardHeader>
                <CardTitle>Events</CardTitle>
              </CardHeader>
              <CardContent>
                <Alert className="m-2">
                  <AlertTitle>Plex</AlertTitle>
                  <AlertDescription>Updated to version 10.1.1</AlertDescription>
                </Alert>
                <Alert className="m-2">
                  <AlertTitle>Transmission</AlertTitle>
                  <AlertDescription>Uninstalled</AlertDescription>
                </Alert>
                <Alert className="m-2">
                  <AlertTitle>UnionFS</AlertTitle>
                  <AlertDescription>Added new filesystem</AlertDescription>
                </Alert>
                <Link href="#" className="hover:underline transition">View events...</Link>
              </CardContent>
            </Card>
            <Card className="m-2">
              <CardHeader>
                <CardTitle>About</CardTitle>
              </CardHeader>
              <CardContent>
                <p><span className="font-bold">Version</span>: 0.0.1</p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </>
  )
}
