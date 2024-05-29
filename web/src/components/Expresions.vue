<script setup>
import ExpresionItem from "@/components/ExpresionItem.vue";

const {expressions} = defineProps(["expressions"])


</script>

<template>
  <!--  {{ expressions }}-->
  <div class="wrapper">
    <ExpresionItem v-for="exp in expressions?.expressions" :key="exp" :expression="exp"
                   :v-if="expressions"></ExpresionItem>
  </div>
</template>

<style lang="scss" scoped>
.wrapper {
  margin: 10px;
}
</style>