<template>
  <v-card outlined rounded="lg">
    <v-toolbar flat color="light-blue lighten-5">
      <v-toolbar-title>{{ attr }} overview</v-toolbar-title>
      <v-spacer />
      <v-btn :to="groupListRoute" small class="primary">View groups</v-btn>
    </v-toolbar>

    <v-container fluid>
      <v-row>
        <v-col>
          <ApiErrorCard v-if="groups.error" :error="groups.error" />
          <PagedGroupsCard
            v-else
            :date-range="dateRange"
            :systems="systems.activeSystems"
            :loading="groups.loading"
            :groups="groups.items"
            :columns="groups.columns"
            :plottable-columns="groups.plottableColumns"
            :plotted-columns="plottedColumns"
            :order="groups.order"
            :axios-params="groups.axiosParams"
          />
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter, useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createQueryEditor, useQueryStore, provideQueryStore, UseUql } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'
import { UseSystems } from '@/tracing/system/use-systems'

// Components
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import PagedGroupsCard from '@/tracing/PagedGroupsCard.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'OverviewAttr',
  components: { ApiErrorCard, PagedGroupsCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()
    const { where } = useQueryStore()

    const attr = computed(() => {
      return route.value.params.attr ?? AttrKey.spanSystem
    })

    const query = computed(() => {
      return createQueryEditor()
        .exploreAttr(attr.value)
        .add(`max(${AttrKey.spanDuration})`)
        .add(where.value)
        .toString()
    })
    provideQueryStore({ query: computed(() => ''), where })

    const groups = useGroups(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...props.systems.axiosParams(),
        query: query.value,
      }
    })

    const plottedColumns = computed(() => {
      return groups.plottableColumns
        .map((col) => col.name)
        .filter((colName) => colName !== `max(${AttrKey.spanDuration})`)
    })

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...props.systems.queryParams(),
          ...groups.order.queryParams(),
          query: query.value,
        },
      }
    })

    useSyncQueryParams({
      fromQuery(queryParams) {
        props.dateRange.parseQueryParams(queryParams)
        props.systems.parseQueryParams(queryParams)
        props.uql.parseQueryParams(queryParams)
        groups.order.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...props.dateRange.queryParams(),
          ...props.systems.queryParams(),
          ...props.uql.queryParams(),
          ...groups.order.queryParams(),
        }
      },
    })

    return {
      AttrKey,

      attr,
      groups,
      plottedColumns,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss"></style>
