import { ref, computed } from 'vue'
import { api, type Feed, type FeedFilter } from '@/services/api'

const feeds = ref<Feed[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const total = ref(0)

export function useFeeds() {
  const fetchFeeds = async (filter?: FeedFilter) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getFeeds(filter)
      if (response.success && response.data) {
        feeds.value = response.data
        total.value = response.meta?.total ?? response.data.length
      } else {
        error.value = response.error?.message ?? 'Failed to fetch feeds'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch feeds'
    } finally {
      loading.value = false
    }
  }

  const fetchFeedsByLocation = async (locationId: string, filter?: FeedFilter) => {
    loading.value = true
    error.value = null

    try {
      const response = await api.getFeedsByLocation(locationId, filter)
      if (response.success && response.data) {
        feeds.value = response.data
        total.value = response.meta?.total ?? response.data.length
      } else {
        error.value = response.error?.message ?? 'Failed to fetch feeds'
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch feeds'
    } finally {
      loading.value = false
    }
  }

  const formattedFeeds = computed(() => {
    return feeds.value.map(feed => ({
      id: feed.id,
      timestamp: formatTimestamp(feed.submitted_at),
      username: feed.username ?? 'anonymous',
      organization: feed.organization ?? '',
      location: feed.location_name ?? '',
      locationId: feed.location_id,
      content: feed.content,
      category: feed.category,
      type: feed.type ?? '',
      coordinates: feed.coordinates,
    }))
  })

  return {
    feeds,
    loading,
    error,
    total,
    formattedFeeds,
    fetchFeeds,
    fetchFeedsByLocation,
  }
}

function formatTimestamp(isoString: string): string {
  const date = new Date(isoString)
  const day = date.getDate().toString().padStart(2, '0')
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const year = date.getFullYear()
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${day}-${month}-${year} ${hours}:${minutes}`
}
