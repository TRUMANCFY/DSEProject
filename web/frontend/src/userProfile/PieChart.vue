<script>
import { Pie } from 'vue-chartjs'

export default {
  extends: Pie,
  props: ['answers', 'data', 'index'],
  data: function() {
    return {
      colors: ['#f87979',  '#41B883','#E46651', '#00D8FF', '#DD1B16'],
    }
  },
  mounted () {
    // this.renderChart(this.data, this.options)

    console.log(this.answers)
    console.log(this.data)

    var length = this.data.length

    this.renderChart({
      labels: this.answers,
      datasets: [
        {
          label: 'GitHub Commits',
          backgroundColor: this.colors.slice(0, length),
          data: this.data
        }
      ]
    })
  }
}
</script>>

