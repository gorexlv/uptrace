<template>
  <v-container fluid>
    <v-row>
      <v-col cols="6">
        <h2 class="mb-5 text-h5">Filtering timeseries</h2>

        <QueryExample query="$metric1 | $metric2">
          <template #description
            >Metric names start with <code>$</code>. Each expr is separated with
            <code>|</code>.</template
          >
        </QueryExample>

        <QueryExample query='$cpu_time{cpu="0",mode="idle"}'>
          <template #description>Select timeseries with given attributes.</template>
        </QueryExample>

        <QueryExample query='$cpu_time{cpu!="0",mode~"user|system"}'>
          <template #description
            >Equal <code>=</code>, not equal <code>!=</code>, regexp match <code>~</code>, regexp no
            match <code>!~</code>.</template
          >
        </QueryExample>

        <QueryExample query="$hits{host.name=localhost} | $misses{host.name=localhost}">
          <template #description>Filter timeseries by <code>host.name</code>.</template>
        </QueryExample>

        <QueryExample query="$hits | $misses | where host.name = 'localhost'">
          <template #description>Filter all timeseries at once by <code>host.name</code>.</template>
        </QueryExample>

        <QueryExample query="$hits | $misses | where host.name exists">
          <template #description
            >Filter timeseries that have <code>host.name</code> attribute.</template
          >
        </QueryExample>
      </v-col>

      <v-col cols="6">
        <h2 class="mb-5 text-h5">Grouping and combining</h2>

        <QueryExample query="$hits | $misses | group by host.name">
          <template #description>Select cache hits and misses on every hostname.</template>
        </QueryExample>

        <QueryExample query="$misses / ($hits + $misses) | group by host.name">
          <template #description>Sum timeseries with matching attributes.</template>
        </QueryExample>

        <QueryExample query='$cpu_time{cpu="0",mode="idle"} as cpu0_idle_time'>
          <template #description>Give timeseries a shorter name (alias).</template>
        </QueryExample>

        <QueryExample
          query="$cache{type=hits} as hits | $cache{type=misses} as misses | hits + misses as total"
        >
          <template #description>Combine timeseries using the aliases.</template>
        </QueryExample>

        <QueryExample query="$metric1 group by service.name | $metric2 group by host.name">
          <template #description>Individual grouping for each timeseries.</template>
        </QueryExample>

        <QueryExample query="$metric_name group by all">
          <template #description>Group by all attributes like Prometheus.</template>
        </QueryExample>
      </v-col>

      <v-col cols="6">
        <h2 class="mb-5 text-h5">Counter Instrument: $cache</h2>

        <QueryExample query="$cache{type=hits}">
          <template #description>Select number of cache hits.</template>
        </QueryExample>

        <QueryExample query="$cache{type=misses}">
          <template #description>Number of cache misses.</template>
        </QueryExample>

        <QueryExample query="$cache{type=hits} + $cache{type=misses}">
          <template #description>Sum of cache hits and misses.</template>
        </QueryExample>

        <QueryExample query="$cache">
          <template #description
            >Sum of cache hits, misses, and possibly other types/timeseries.</template
          >
        </QueryExample>

        <QueryExample query="per_min($cache) | per_sec($cache)">
          <template #description> Number of cache operations per minute and per second. </template>
        </QueryExample>

        <QueryExample query="sum($cache) / _minutes | sum($cache) / _seconds">
          <template #description>The same as the previous query.</template>
        </QueryExample>
      </v-col>

      <v-col cols="6">
        <h2 class="mb-5 text-h5">Histogram Instrument: $srv_duration</h2>

        <QueryExample query="p50($srv_duration)">
          <template #description>P50 duration.</template>
        </QueryExample>

        <QueryExample query="p90($srv_duration{env=prod}) | p90($srv_duration{env=dev})">
          <template #description
            >P90 duration in <code>prod</code> and <code>env</code> environments.</template
          >
        </QueryExample>

        <QueryExample query="avg($srv_duration) group by host.name">
          <template #description>Avg duration on each hostname.</template>
        </QueryExample>

        <QueryExample query='avg($srv_duration{host.name~"api\d+$"})'>
          <template #description>Avg duration on hostnames matching the regexp.</template>
        </QueryExample>

        <QueryExample query="per_min(count($srv_duration)) | per_sec(count($srv_duration))">
          <template #description>Number of requests per minute and per second.</template>
        </QueryExample>

        <QueryExample query="min($srv_duration) | max($srv_duration)">
          <template #description>Min and max duration.</template>
        </QueryExample>
      </v-col>

      <v-col cols="6">
        <h2 class="mb-5 text-h5">Advanced</h2>

        <QueryExample query="uniq($hits, host.name, service.name) as num_timeseries">
          <template #description
            >Count the number of unique combinations of <code>host.name</code> and
            <code>service.name</code>.</template
          >
        </QueryExample>

        <QueryExample query="delta($kafka_part_offset) as messages_processed">
          <template #description
            >Calculate the difference between the current and previous values.</template
          >
        </QueryExample>

        <QueryExample query="$load_avg_15m / uniq($cpu_time, cpu) as cpu_util">
          <template #description
            >Calculate CPU utilization using <code>system.cpu.load_average.15m</code> and
            <code>system.cpu.time</code>.</template
          >
        </QueryExample>

        <QueryExample query="min($cache..time), max($cache..time)">
          <template #description
            >Get the first/last time the metric received an update (note the double dot)</template
          >
        </QueryExample>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Components
import QueryExample from '@/components/QueryExample.vue'

export default defineComponent({
  name: 'CheatSheetCard',
  components: { QueryExample },
})
</script>

<style lang="scss" scoped></style>
