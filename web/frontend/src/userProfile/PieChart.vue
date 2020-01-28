<script>
import { Pie } from 'vue-chartjs'

export default {
  extends: Pie,
  props: ['answers', 'data', 'index'],
  mounted () {
    // this.renderChart(this.data, this.options)

    console.log(this.answers)
    console.log(this.data)

    this.renderChart({
      labels: this.answers,
      datasets: [
        {
          label: 'GitHub Commits',
          backgroundColor: '#f87979',
          data: this.data
        }
      ]
    })
  }
}
</script>>

