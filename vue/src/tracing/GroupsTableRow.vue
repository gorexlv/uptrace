<template>
  <tr :class="{ 'cursor-pointer': !hideActions }" @click="!hideActions && expand(!isExpanded)">
    <td v-if="headerValues.includes('_name')" class="word-break-all">
      {{ group[AttrKey.displayName] }}
    </td>
    <td v-for="col in plainColumns" :key="col.name">
      <AnyValue :value="group[col.name]" :unit="col.unit" />
    </td>
    <td v-if="headerValues.includes(AttrKey.spanSystem)">
      <router-link :to="systemRoute" @click.stop.prevent>
        {{ group[AttrKey.spanSystem] }}
      </router-link>
    </td>
    <td v-for="col in plottableColumns" :key="col.name">
      <div class="d-flex align-center">
        <SparklineChart
          v-if="plottedColumns.includes(col.name)"
          :name="col.name"
          :line="timeseries.data[col.name] ?? []"
          :time="timeseries.time"
          :unit="col.unit"
          :color="columnMap[col.name].color"
          :group="group._id"
          class="mr-2"
        />
        <NumValue :value="group[col.name]" :name="col.name" :unit="col.unit" />
      </div>
    </td>
    <td v-if="!hideActions" class="text-right text-no-wrap">
      <NewMonitorMenu
        :name="group._name"
        :axios-params="axiosParams"
        :where="group._query"
        :events-mode="eventsMode"
      >
        <template #header-item>
          <slot name="summary-item" :group="group" :is-event="eventsMode" />
        </template>
      </NewMonitorMenu>

      <v-btn icon title="Filter spans for this group" :to="itemListRoute" @click.native.stop>
        <v-icon>mdi-filter-variant</v-icon>
      </v-btn>

      <v-btn v-if="isExpanded" icon title="Hide spans" @click.stop="expand(false)">
        <v-icon size="30">mdi-chevron-up</v-icon>
      </v-btn>
      <v-btn v-else icon title="View spans" @click.stop="expand(true)">
        <v-icon size="30">mdi-chevron-down</v-icon>
      </v-btn>
    </td>
  </tr>
</template>

<script lang="ts">
import { omit, truncate } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { createQueryEditor, joinQuery, injectQueryStore } from '@/use/uql'
import { useGroupTimeseries, Group, ColumnInfo } from '@/tracing/use-explore-spans'

// Components
import SparklineChart from '@/components/SparklineChart.vue'
import NewMonitorMenu from '@/tracing/NewMonitorMenu.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { MetricColumn } from '@/metrics/types'

export default defineComponent({
  name: 'GroupsTableRow',
  components: { SparklineChart, NewMonitorMenu },

  props: {
    systems: {
      type: Array as PropType<string[]>,
      required: true,
    },
    query: {
      type: String,
      default: '',
    },
    groupingColumns: {
      type: Array as PropType<string[]>,
      required: true,
    },
    plainColumns: {
      type: Array as PropType<string[]>,
      required: true,
    },
    plottableColumns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    plottedColumns: {
      type: Array as PropType<string[]>,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      required: true,
    },
    axiosParams: {
      type: Object,
      default: undefined,
    },

    headers: {
      type: Array,
      required: true,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      required: true,
    },
    group: {
      type: Object as PropType<Group>,
      required: true,
    },
    hideActions: {
      type: Boolean,
      default: false,
    },
    isExpanded: {
      type: Boolean,
      required: true,
    },
    expand: {
      type: Function,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()
    const { query, where } = injectQueryStore()

    const timeseries = useGroupTimeseries(() => {
      if (!props.plottedColumns.length) {
        return undefined
      }

      const query = joinQuery([props.axiosParams.query, props.group._query])
      return {
        ...props.axiosParams,
        query,
        column: props.plottedColumns,
      }
    })

    const headerValues = computed((): string[] => {
      return props.headers.map((header: any) => header.value)
    })

    const systemRoute = computed(() => {
      const query = createQueryEditor().exploreAttr(AttrKey.spanGroupId).add(where.value).toString()
      return {
        name: 'SpanGroupList',
        query: {
          ...omit(route.value.query, 'columns'),
          query: query,
          system: props.group[AttrKey.spanSystem],
        },
      }
    })

    const itemListRoute = computed(() => {
      const editor = query.value
        ? createQueryEditor(query.value)
        : createQueryEditor().exploreAttr(AttrKey.spanGroupId, props.eventsMode)
      editor.add(props.group._query)

      for (let colName of props.groupingColumns) {
        const value = props.group[colName]
        editor.where(colName, '=', value)
      }

      const queryParams: Record<string, any> = {
        ...route.value.query,
        query: editor.toString(),
      }

      const system = props.group[AttrKey.spanSystem]
      if (system) {
        queryParams.system = system
      }

      return {
        name: 'SpanList',
        query: queryParams,
      }
    })

    function systemsForGroup(): string[] {
      const system = props.group[AttrKey.spanSystem]
      if (system) {
        return [system]
      }
      return props.systems
    }

    return {
      AttrKey,

      timeseries,
      headerValues,

      systemRoute,
      itemListRoute,

      systemsForGroup,
      truncate,
    }
  },
})
</script>

<style lang="scss" scoped></style>
