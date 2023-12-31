"use client"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { NavigationMenu, NavigationMenuContent, NavigationMenuItem, NavigationMenuLink, NavigationMenuList, NavigationMenuTrigger, navigationMenuTriggerStyle } from "@/components/ui/navigation-menu"
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"
import {useState, useEffect} from "react"
import {login} from "@/lib/auth"

const apps = [
  {
    name: "qBittorrent",
    slug_name: "qbittorrent",
    status: "running",
  },
  {
    name: "Plex",
    slug_name: "plex",
    status: "running",
  },
  {
    name: "Rclone",
    slug_name: "rclone",
    status: "running",
  },
  {
    name: "Radarr",
    slug_name: "radarr",
    status: "running",
  },
  {
    name: "Sonarr",
    slug_name: "sonarr",
    status: "running",
  },
  {
    name: "Jackett",
    slug_name: "jackett",
    status: "running",
  },
  {
    name: "Overseerr",
    slug_name: "overseerr",
    status: "running",
  }
]

export default function Navbar() {
    const [ loggedIn, setLoggedIn ] = useState(false)
    useEffect(() => {
        const token = localStorage.getItem("token")
        if (token) {
          setLoggedIn(true)
        }
    }, [])

    function logout(e: any) {
      e.preventDefault()
      localStorage.removeItem("token")
      setLoggedIn(false)
    }

    async function test_login(e: any) {
      e.preventDefault()
      login("reece_jones@icloud.com", "password")
      setLoggedIn(true)
    }
    return (
      <div className="w-full flex border-b-2">
        <NavigationMenu className="p-4">
          <NavigationMenuList>
            <NavigationMenuItem>
              <Link href="/" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Home
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/dashboard" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Dashboard
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/apps" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Apps
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/tools" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Tools
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/settings" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Settings
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>
        <div className="ml-auto p-4">
          <DropdownMenu>
            <DropdownMenuTrigger>
              <Avatar>
                <AvatarImage src="https://github.com/ReeceJones.png"/>
                <AvatarFallback>RJ</AvatarFallback>
              </Avatar>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-60">
              {
                loggedIn ? (
                  <>
                    <DropdownMenuItem>
                      <Link href="/settings">Settings</Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Link href="/logout" onClick={logout}>Logout</Link>
                    </DropdownMenuItem>
                  </>
                ) : (
                  <>
                    <DropdownMenuItem>
                      <Link href="/login" onClick={test_login}>Login</Link>
                    </DropdownMenuItem>
                  </>
                )
              }
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    )
}