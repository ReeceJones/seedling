"use client"
import { useState, useEffect } from "react"
import { getServices, installService, Service } from "@/lib/services"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useToast } from "@/components/ui/use-toast"
import Link from "next/link"

// const apps = [
//     {
//         name: "qBittorrent",
//         slug_name: "qbittorrent",
//         status: "running",
//         url: "http://localhost:8080",
//         description: "qBittorrent is a free and open-source BitTorrent client, available for Linux, FreeBSD, macOS and Windows."
//     },
//     {
//         name: "Plex",
//         slug_name: "plex",
//         status: "running",
//         url: "http://localhost:32400",
//         description: "Plex is a client-server media player system and software suite comprising two main components."
//     },
//     {
//         name: "Rclone",
//         slug_name: "rclone",
//         status: "running",
//         url: "http://localhost:5572",
//         description: "Rclone is a command-line program to manage files on cloud storage."
//     },
//     {
//         name: "Radarr",
//         slug_name: "radarr",
//         status: "running",
//         url: "http://localhost:7878",
//         description: "Radarr is a movie collection manager for Usenet and BitTorrent users."
//     },
//     {
//         name: "Sonarr",
//         slug_name: "sonarr",
//         status: "running",
//         url: "http://localhost:8989",
//         description: "Sonarr is a PVR for Usenet and BitTorrent users."
//     },
//     {
//         name: "Jackett",
//         slug_name: "jackett",
//         status: "running",
//         url: "http://localhost:9117",
//         description: "Jackett works as a proxy server: it translates queries from apps (Sonarr, Radarr, SickRage, CouchPotato, Mylar, Lidarr, DuckieTV, qBittorrent, Nefarious etc) into tracker-site-specific http queries, parses the html response, then sends results back to the requesting software."
//     },
//     {
//         name: "Overseerr",
//         slug_name: "overseerr",
//         status: "running",
//         url: "http://localhost:5055",
//         description: "Overseerr is a request management and media discovery tool for the Plex ecosystem."
//     },
//     {
//         name: "Tautulli",
//         slug_name: "tautulli",
//         status: "not_installed",
//         url: "http://localhost:8181",
//         description: "Tautulli is a 3rd party application that you can run alongside your Plex Media Server to monitor activity and track various statistics."
//     },
//     {
//         name: "Jellyfin",
//         slug_name: "jellyfin",
//         status: "not_installed",
//         url: "http://localhost:8096",
//         description: "Jellyfin is the volunteer-built media solution that puts you in control of your media. Stream to any device from your own server, with no strings attached. Your media, your server, your way."
//     },
//     {
//         name: "Prowlarr",
//         slug_name: "prowlarr",
//         status: "not_installed",
//         url: "http://localhost:9696",
//         description: "Prowlarr is a indexer manager/proxy built on the popular arr .net/reactjs base stack to integrate with your various PVR apps."
//     },
//     {
//         name: "Lidarr",
//         slug_name: "lidarr",
//         status: "not_installed",
//         url: "http://localhost:8686",
//         description: "Lidarr is a music collection manager for Usenet and BitTorrent users."
//     },
//     {
//         name: "NZBGet",
//         slug_name: "nzbget",
//         status: "not_installed",
//         url: "http://localhost:6789",
//         description: "NZBGet is a binary downloader, which downloads files from Usenet based on information given in nzb-files."
//     },
//     {
//         name: "NZBHydra2",
//         slug_name: "nzbhydra2",
//         status: "not_installed",
//         url: "http://localhost:5076",
//         description: "NZBHydra2 is a meta search for NZB indexers."
//     },
//     {
//         name: "NZBHydra",
//         slug_name: "nzbhydra",
//         status: "not_installed",
//         url: "http://localhost:5075",
//         description: "NZBHydra is a meta search for NZB indexers."
//     },
//     {
//         name: "NZBGet",
//         slug_name: "nzbget",
//         status: "not_installed",
//         url: "http://localhost:6789",
//         description: "NZBGet is a binary downloader, which downloads files from Usenet based on information given in nzb-files."
//     }
// ]

export default function Home() {
  const [ services, setServices ] = useState<Service[]>([])
  const { toast } = useToast()

  useEffect(() => {
    getServices().then((data) => {
      setServices(data)
    })
  }, [])

  async function installServiceAsUser(service: Service) {
    toast({
      title: service.name,
      description: "Installing app...",
      duration: 300_000,
    })
    try {
      await installService(service.key)
      toast({
        title: service.name,
        description: "Installed app!",
        duration: 300_000,
      })
    }
    catch (e) {
      console.log(e)
      toast({
        title: service.name,
        description: "Failed to install app!",
        duration: 300_000,
      })
    }

    setServices(await getServices())
  }

  return (
    <>
      <div className="w-full p-10">
        <div className="mx-auto w-[1200px]">
          <div className="m-4">
            <h1 className="text-4xl font-bold">Apps</h1>
          </div>
          <div className="grid grid-cols-2 w-[1200px]">
            {
              services.map((app) => (
                <div className="col-span-1 m-2">
                  <Card className="h-full">
                    <CardHeader>
                      <CardTitle><span className="align-middle">{app.name}</span> <Badge className="float-right">{app.status}</Badge></CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p>{app.description}</p>
                      <br/>
                      <Link href={app.project_url}>
                        {app.project_url}
                      </Link>
                    </CardContent>
                    <CardFooter>
                      {
                        app.is_running || app.is_installed ? (
                          <Button>Settings</Button>
                        ) : (
                          <Button onClick={async () => await installServiceAsUser(app)}>Install</Button>
                        )
                      }
                    </CardFooter>
                  </Card>
                </div>
              ))
            }
          </div>
        </div>
      </div>
    </>
  )
  }
  