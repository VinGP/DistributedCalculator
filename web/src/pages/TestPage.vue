<script setup>

import Expresions from "@/components/Expresions.vue";
</script>

<template>
  <body>
  <header contentEditable>Header</header>
  <main contentEditable>
    <div class="content">
      Contant
    </div>
    <div>
      <Expresions/>
    </div>
  </main>
  <footer contentEditable>Footer</footer>
  </body>
</template>

<style lang="scss" scoped>
body {
  min-height: 100vh;
  margin: 0;
}

header {
  min-height: 50px;
  background: lightcyan;
}

footer {
  min-height: 50px;
  background: PapayaWhip;
}


/* The article fills all the space between header & footer */

body {
  display: flex;
  flex-direction: column;
}

.content {
  min-height: calc(100vh - 100px);
  //height: 1000px;
}

main {
  //height: 1000px;
  background: red;
  //flex: 1;
}
</style>