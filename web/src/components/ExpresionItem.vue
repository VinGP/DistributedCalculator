<script setup>
const {expression} = defineProps(["expression"])


var statusColor = "#4d9efe"

if (expression?.error){
  statusColor = "red"
} else if (expression.result) {
  statusColor = "green"
}


</script>

<template>

<!--  <div>{{ expression }}</div>-->
  <div class="container">
    <div class="status" :style="{'background': statusColor}"></div>
    <div class="expression text">
      {{ expression.expression }} <span v-if="expression && !expression?.error">= {{ expression.result ? expression.result : "?" }} </span>
    </div>
    <div v-if="expression.error" class="error text" >({{ expression.error }})</div>
<!--    <div class="text" :style="{'margin-left': '10px'}">{{expression.id}}</div>-->
  </div>

</template>

<style lang="scss" scoped>
.container {
  display: flex;
  align-items: center;
  //margin: 5px;
}
.expression{
  margin-left: 10px;
}

.status{
  height: 30px;
  width: 30px;
  border-radius: 100%;
  background: #4d9efe;
}

.text {
  font-size: 25px;
}
.error{
  margin-left: 10px;
  color: red;
}


</style>