<template>
  <span>
    <span class="text-no-wrap">
      <DateValue :value="around" format="short" class="mr-2 text-subtitle-2" />
      <PeriodPickerMenu
        v-if="dateRange.duration"
        :value="dateRange.duration"
        :periods="periods"
        @input="onInputPeriod"
      />
    </span>
    <v-btn v-if="showReload" small outlined class="ml-2" @click="dateRange.reload()">
      <v-icon small left>mdi-refresh</v-icon>
      <span>Reload</span>
    </v-btn>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, onMounted, watch, watchEffect, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import PeriodPickerMenu from '@/components/date/PeriodPickerMenu.vue'

// Utilities
import { periodsForDays } from '@/models/period'
import { HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'FixedDateRangePicker',
  components: {
    PeriodPickerMenu,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    around: {
      type: String,
      required: true,
    },
    rangeDays: {
      type: Number,
      default: 10,
    },
    showReload: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const periods = computed(() => {
      return periodsForDays(props.rangeDays)
    })

    onMounted(() => {
      watchEffect(() => {
        if (props.dateRange.duration) {
          return
        }

        const period = periods.value.find((p) => p.milliseconds === HOUR)
        if (period) {
          props.dateRange.changeAround(props.around, period.milliseconds)
          return
        }

        props.dateRange.changeAround(props.around, periods.value[0].milliseconds)
      })
    })

    watch(
      () => props.around,
      (date) => {
        props.dateRange.changeAround(date)
      },
      { immediate: true },
    )

    function onInputPeriod(ms: number) {
      props.dateRange.changeAround(props.around, ms)
    }

    return { periods, onInputPeriod }
  },
})
</script>

<style lang="scss" scoped></style>
