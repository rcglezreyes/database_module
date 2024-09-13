import { Box } from '@mui/material';
import { BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid, Cell } from 'recharts';

const AssessmentByIdBarChartComponent = ({ data }) => {
  
  const threshold = data.reduce((acc, item) => acc + item.StudentCount, 0) / data.length;

  // Función para determinar el color de cada barra
  const getBarColor = (studentCount) => {
    return studentCount >= threshold ? "#82ca9d" : "#8884d8";
  };

  return (
    <Box 
      p={2} 
      sx={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100%', // Ocupa toda la altura visible
        width: '100%',
      }}
    >
      <BarChart 
        width={window.innerWidth * 0.4}  // Ancho del 90% de la ventana
        height={window.innerHeight * 0.4} // Alto del 80% de la ventana
        data={data}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="AssessmentID" />
        <YAxis />
        <Tooltip />
        <Bar dataKey="StudentCount">
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={getBarColor(entry.StudentCount)} />
          ))}
        </Bar>
      </BarChart>
    </Box>
  );
};

export default AssessmentByIdBarChartComponent;
