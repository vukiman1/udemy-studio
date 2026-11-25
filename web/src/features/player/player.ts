import { createPlayer } from '@videojs/react'
import { videoFeatures } from '@videojs/react/video'

export const { Player, usePlayer } = createPlayer({ features: videoFeatures })
