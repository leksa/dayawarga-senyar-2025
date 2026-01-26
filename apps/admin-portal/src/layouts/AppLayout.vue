<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import {
  LayoutDashboard,
  Building2,
  Users,
  UserCircle,
  Menu,
  X,
  LogOut,
  ChevronDown,
  FolderCheck,
  UserCog,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const isSidebarOpen = ref(true)
const isMobileSidebarOpen = ref(false)

// Navigation items - filtered based on user role
const navItems = computed(() => {
  const items = [
    { name: 'Dashboard', path: '/', icon: LayoutDashboard },
    { name: 'Organisasi', path: '/organizations', icon: Building2 },
    { name: 'Tim/Grup', path: '/groups', icon: Users },
    { name: 'Relawan', path: '/relawan', icon: UserCircle },
  ]

  // Project requests: visible to all (org_admin can request, super_admin can manage)
  items.push({ name: 'Permintaan ODK', path: '/project-requests', icon: FolderCheck })

  // Users management: only visible to super_admin
  if (authStore.isSuperAdmin) {
    items.push({ name: 'Pengguna', path: '/users', icon: UserCog })
  }

  return items
})

const currentNavItem = computed(() => {
  return navItems.value.find((item) => route.path === item.path || route.path.startsWith(item.path + '/'))
})

const userInitials = computed(() => {
  const name = authStore.displayName
  if (!name) return 'U'
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
})

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

function toggleMobileSidebar() {
  isMobileSidebarOpen.value = !isMobileSidebarOpen.value
}

async function handleLogout() {
  await authStore.logout()
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <!-- Mobile sidebar backdrop -->
    <Transition
      enter-active-class="transition-opacity duration-300"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-300"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isMobileSidebarOpen"
        class="fixed inset-0 z-40 bg-black/50 lg:hidden"
        @click="toggleMobileSidebar"
      />
    </Transition>

    <!-- Sidebar -->
    <aside
      :class="[
        'fixed left-0 top-0 z-50 h-screen bg-sidebar border-r border-sidebar-border transition-all duration-300',
        isSidebarOpen ? 'w-64' : 'w-16',
        isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
      ]"
    >
      <!-- Logo -->
      <div class="flex h-16 items-center justify-between px-4 border-b border-sidebar-border">
        <RouterLink to="/" class="flex items-center gap-3">
          <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
            DW
          </div>
          <span
            v-if="isSidebarOpen"
            class="font-semibold text-sidebar-foreground whitespace-nowrap"
          >
            Dayawarga
          </span>
        </RouterLink>
        <Button
          v-if="isSidebarOpen"
          variant="ghost"
          size="icon"
          class="lg:hidden text-sidebar-foreground"
          @click="toggleMobileSidebar"
        >
          <X class="h-5 w-5" />
        </Button>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 space-y-1 p-3">
        <RouterLink
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          :class="[
            'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
            route.path === item.path || route.path.startsWith(item.path + '/')
              ? 'bg-sidebar-accent text-sidebar-accent-foreground'
              : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground',
          ]"
          @click="isMobileSidebarOpen && toggleMobileSidebar()"
        >
          <component :is="item.icon" class="h-5 w-5 shrink-0" />
          <span v-if="isSidebarOpen" class="whitespace-nowrap">{{ item.name }}</span>
        </RouterLink>
      </nav>

      <!-- Sidebar footer - collapse button (desktop only) -->
      <div class="hidden lg:block border-t border-sidebar-border p-3">
        <Button
          variant="ghost"
          size="sm"
          class="w-full justify-start text-sidebar-foreground/70 hover:text-sidebar-foreground"
          @click="toggleSidebar"
        >
          <Menu class="h-5 w-5" />
          <span v-if="isSidebarOpen" class="ml-3">Collapse</span>
        </Button>
      </div>
    </aside>

    <!-- Main content -->
    <div
      :class="[
        'flex flex-col min-h-screen transition-all duration-300',
        isSidebarOpen ? 'lg:ml-64' : 'lg:ml-16',
      ]"
    >
      <!-- Header -->
      <header class="sticky top-0 z-30 flex h-16 items-center justify-between border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 lg:px-6">
        <div class="flex items-center gap-4">
          <!-- Mobile menu button -->
          <Button
            variant="ghost"
            size="icon"
            class="lg:hidden"
            @click="toggleMobileSidebar"
          >
            <Menu class="h-5 w-5" />
          </Button>

          <!-- Page title -->
          <div>
            <h1 class="text-lg font-semibold">
              {{ currentNavItem?.name || 'Dashboard' }}
            </h1>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <ThemeToggle />

          <Separator orientation="vertical" class="h-6" />

          <!-- User menu -->
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" class="gap-2 pl-2 pr-1">
                <Avatar class="h-8 w-8">
                  <AvatarFallback class="bg-primary/10 text-primary text-xs font-medium">
                    {{ userInitials }}
                  </AvatarFallback>
                </Avatar>
                <span class="hidden sm:inline-block text-sm font-medium">
                  {{ authStore.displayName }}
                </span>
                <ChevronDown class="h-4 w-4 text-muted-foreground" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-48">
              <div class="px-2 py-1.5">
                <p class="text-sm font-medium">{{ authStore.displayName }}</p>
                <p class="text-xs text-muted-foreground">{{ authStore.profile?.email }}</p>
                <p v-if="authStore.userRole" class="text-xs text-primary font-medium mt-0.5">
                  {{ authStore.isSuperAdmin ? 'Super Admin' : authStore.isOrgAdmin ? 'Admin Organisasi' : 'Member' }}
                </p>
              </div>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="handleLogout" class="text-destructive focus:text-destructive">
                <LogOut class="mr-2 h-4 w-4" />
                Keluar
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 p-4 lg:p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
