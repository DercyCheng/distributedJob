#!/usr/bin/env python3
"""
Go-Job MCP Server
提供智能任务调度的MCP工具集成
"""

import asyncio
import json
import sys
import logging
from typing import Dict, List, Any, Optional
from datetime import datetime, timedelta
import requests
import os
from dashscope import Generation

# MCP Server Implementation
class GoJobMCPServer:
    def __init__(self):
        self.api_base_url = os.getenv('GO_JOB_API_URL', 'http://localhost:8080/api/v1')
        self.auth_token = os.getenv('GO_JOB_AUTH_TOKEN', '')
        self.dashscope_api_key = os.getenv('DASHSCOPE_API_KEY', '')
        
        # Configure logging
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)
        
        # Initialize tools
        self.tools = {
            'list_jobs': self.list_jobs,
            'analyze_job_performance': self.analyze_job_performance,
            'optimize_schedule': self.optimize_schedule,
            'predict_resource_usage': self.predict_resource_usage,
            'get_recommendations': self.get_recommendations,
            'create_smart_schedule': self.create_smart_schedule,
        }

    def _make_request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict:
        """Make HTTP request to Go-Job API"""
        url = f"{self.api_base_url}{endpoint}"
        headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {self.auth_token}' if self.auth_token else ''
        }
        
        try:
            if method.upper() == 'GET':
                response = requests.get(url, headers=headers, params=data)
            elif method.upper() == 'POST':
                response = requests.post(url, headers=headers, json=data)
            elif method.upper() == 'PUT':
                response = requests.put(url, headers=headers, json=data)
            elif method.upper() == 'DELETE':
                response = requests.delete(url, headers=headers)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
            
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            self.logger.error(f"API request failed: {e}")
            return {"error": str(e)}

    def _call_dashscope_api(self, prompt: str, context: Dict = None) -> str:
        """Call DashScope API for AI analysis"""
        if not self.dashscope_api_key:
            return "DashScope API密钥未配置"
        
        try:
            # Prepare messages
            messages = [
                {
                    "role": "system",
                    "content": "你是一个专业的任务调度系统AI助手，专门用于分析和优化定时任务的执行。请提供专业、详细的分析建议。"
                },
                {
                    "role": "user", 
                    "content": f"上下文：{json.dumps(context, ensure_ascii=False) if context else ''}\n\n问题：{prompt}"
                }
            ]
            
            response = Generation.call(
                model='qwen-max',
                messages=messages,
                temperature=0.7,
                api_key=self.dashscope_api_key
            )
            
            if response.status_code == 200:
                return response.output.text
            else:
                return f"AI分析失败: {response.message}"
                
        except Exception as e:
            self.logger.error(f"DashScope API call failed: {e}")
            return f"AI分析异常: {str(e)}"

    async def list_jobs(self, args: Dict) -> Dict:
        """列出系统中的任务"""
        department_id = args.get('department_id', '')
        status = args.get('status', '')
        limit = args.get('limit', 10)
        
        params = {
            'limit': limit
        }
        if department_id:
            params['department_id'] = department_id
        if status:
            params['status'] = status
            
        result = self._make_request('GET', '/jobs', params)
        
        if 'error' in result:
            return {"success": False, "error": result['error']}
        
        jobs = result.get('jobs', [])
        summary = {
            "total_jobs": len(jobs),
            "active_jobs": len([j for j in jobs if j.get('enabled', True)]),
            "departments": list(set([j.get('department_id', '') for j in jobs if j.get('department_id')])),
        }
        
        return {
            "success": True,
            "data": {
                "jobs": jobs,
                "summary": summary
            }
        }

    async def analyze_job_performance(self, args: Dict) -> Dict:
        """分析任务执行性能"""
        job_id = args.get('job_id')
        days = args.get('days', 7)
        metric = args.get('metric', 'all')
        
        if not job_id:
            return {"success": False, "error": "job_id is required"}
        
        # 获取任务信息
        job_result = self._make_request('GET', f'/jobs/{job_id}')
        if 'error' in job_result:
            return {"success": False, "error": job_result['error']}
        
        # 获取执行历史
        exec_result = self._make_request('GET', f'/jobs/{job_id}/executions', {
            'days': days
        })
        
        executions = exec_result.get('executions', [])
        
        # 计算性能指标
        total_executions = len(executions)
        success_count = len([e for e in executions if e.get('status') == 'success'])
        failed_count = len([e for e in executions if e.get('status') == 'failed'])
        timeout_count = len([e for e in executions if e.get('status') == 'timeout'])
        
        success_rate = (success_count / total_executions * 100) if total_executions > 0 else 0
        
        # 计算平均执行时间
        durations = []
        for exec in executions:
            if exec.get('started_at') and exec.get('finished_at'):
                start = datetime.fromisoformat(exec['started_at'].replace('Z', '+00:00'))
                end = datetime.fromisoformat(exec['finished_at'].replace('Z', '+00:00'))
                durations.append((end - start).total_seconds())
        
        avg_duration = sum(durations) / len(durations) if durations else 0
        
        # 构建分析上下文
        analysis_context = {
            "job_info": job_result,
            "performance_metrics": {
                "total_executions": total_executions,
                "success_rate": round(success_rate, 2),
                "failed_count": failed_count,
                "timeout_count": timeout_count,
                "avg_duration_seconds": round(avg_duration, 2)
            },
            "recent_executions": executions[:10]  # 最近10次执行
        }
        
        # 调用AI分析
        ai_analysis = self._call_dashscope_api(
            f"请分析这个定时任务在过去{days}天的性能表现，并提供优化建议。重点关注{metric}指标。",
            analysis_context
        )
        
        return {
            "success": True,
            "data": {
                "job_id": job_id,
                "analysis_period_days": days,
                "metrics": analysis_context["performance_metrics"],
                "ai_analysis": ai_analysis,
                "recommendations": self._generate_performance_recommendations(analysis_context)
            }
        }

    async def optimize_schedule(self, args: Dict) -> Dict:
        """优化任务调度时间"""
        job_ids = args.get('job_ids', [])
        goal = args.get('goal', 'load_balance')
        
        if not job_ids:
            return {"success": False, "error": "job_ids is required"}
        
        # 获取任务信息
        jobs_data = []
        for job_id in job_ids:
            job_result = self._make_request('GET', f'/jobs/{job_id}')
            if 'error' not in job_result:
                jobs_data.append(job_result)
        
        if not jobs_data:
            return {"success": False, "error": "No valid jobs found"}
        
        # 获取系统负载数据
        stats_result = self._make_request('GET', '/stats/dashboard')
        system_stats = stats_result.get('data', {})
        
        # 构建优化上下文
        optimization_context = {
            "jobs": jobs_data,
            "optimization_goal": goal,
            "system_stats": system_stats,
            "current_time": datetime.now().isoformat()
        }
        
        # 调用AI进行调度优化
        ai_optimization = self._call_dashscope_api(
            f"请为这些任务制定最优的调度时间安排，优化目标是{goal}。考虑系统负载、任务依赖关系和执行时间分布。",
            optimization_context
        )
        
        return {
            "success": True,
            "data": {
                "optimization_goal": goal,
                "analyzed_jobs": len(jobs_data),
                "ai_recommendations": ai_optimization,
                "suggested_schedules": self._generate_schedule_suggestions(jobs_data, goal)
            }
        }

    async def predict_resource_usage(self, args: Dict) -> Dict:
        """预测系统资源使用情况"""
        hours = args.get('hours', 24)
        granularity = args.get('granularity', 'hour')
        
        # 获取历史统计数据
        stats_result = self._make_request('GET', '/stats/dashboard')
        worker_stats = self._make_request('GET', '/stats/workers')
        job_stats = self._make_request('GET', '/stats/jobs')
        
        # 构建预测上下文
        prediction_context = {
            "prediction_period_hours": hours,
            "granularity": granularity,
            "current_stats": stats_result.get('data', {}),
            "worker_stats": worker_stats.get('data', {}),
            "job_stats": job_stats.get('data', {}),
            "timestamp": datetime.now().isoformat()
        }
        
        # 调用AI进行资源使用预测
        ai_prediction = self._call_dashscope_api(
            f"基于历史数据预测未来{hours}小时的系统资源使用情况，预测粒度为{granularity}。",
            prediction_context
        )
        
        return {
            "success": True,
            "data": {
                "prediction_period_hours": hours,
                "granularity": granularity,
                "ai_prediction": ai_prediction,
                "resource_trends": self._analyze_resource_trends(prediction_context)
            }
        }

    async def get_recommendations(self, args: Dict) -> Dict:
        """获取任务优化建议"""
        job_id = args.get('job_id')
        recommendation_type = args.get('type', 'all')
        
        if job_id:
            # 单个任务的建议
            job_result = self._make_request('GET', f'/jobs/{job_id}')
            if 'error' in job_result:
                return {"success": False, "error": job_result['error']}
            
            exec_result = self._make_request('GET', f'/jobs/{job_id}/executions', {'limit': 20})
            
            context = {
                "job": job_result,
                "recent_executions": exec_result.get('executions', []),
                "recommendation_type": recommendation_type
            }
        else:
            # 全系统建议
            jobs_result = self._make_request('GET', '/jobs', {'limit': 100})
            stats_result = self._make_request('GET', '/stats/dashboard')
            
            context = {
                "jobs": jobs_result.get('jobs', []),
                "system_stats": stats_result.get('data', {}),
                "recommendation_type": recommendation_type
            }
        
        # 调用AI生成建议
        ai_recommendations = self._call_dashscope_api(
            f"请提供{recommendation_type}类型的任务优化建议，重点关注性能、可靠性和成本优化。",
            context
        )
        
        return {
            "success": True,
            "data": {
                "recommendation_type": recommendation_type,
                "target": "single_job" if job_id else "system",
                "ai_recommendations": ai_recommendations,
                "actionable_items": self._extract_actionable_items(ai_recommendations)
            }
        }

    async def create_smart_schedule(self, args: Dict) -> Dict:
        """创建智能调度方案"""
        job_requirements = args.get('requirements', {})
        constraints = args.get('constraints', {})
        
        # 获取系统当前状态
        stats_result = self._make_request('GET', '/stats/dashboard')
        jobs_result = self._make_request('GET', '/jobs')
        
        # 构建智能调度上下文
        context = {
            "requirements": job_requirements,
            "constraints": constraints,
            "system_state": stats_result.get('data', {}),
            "existing_jobs": jobs_result.get('jobs', [])
        }
        
        # 调用AI生成智能调度方案
        ai_schedule = self._call_dashscope_api(
            "基于系统当前状态和需求约束，创建一个智能的任务调度方案。",
            context
        )
        
        return {
            "success": True,
            "data": {
                "schedule_context": context,
                "ai_generated_schedule": ai_schedule,
                "implementation_steps": self._generate_implementation_steps(ai_schedule)
            }
        }

    def _generate_performance_recommendations(self, context: Dict) -> List[str]:
        """生成性能优化建议"""
        recommendations = []
        metrics = context.get("performance_metrics", {})
        
        if metrics.get("success_rate", 0) < 95:
            recommendations.append("成功率偏低，建议检查任务执行环境和错误处理机制")
        
        if metrics.get("avg_duration_seconds", 0) > 300:
            recommendations.append("平均执行时间较长，建议优化任务逻辑或增加超时设置")
        
        if metrics.get("timeout_count", 0) > 0:
            recommendations.append("存在超时任务，建议调整超时时间或优化任务性能")
        
        return recommendations

    def _generate_schedule_suggestions(self, jobs_data: List[Dict], goal: str) -> List[Dict]:
        """生成调度建议"""
        suggestions = []
        
        for job in jobs_data:
            suggestion = {
                "job_id": job.get("id"),
                "current_cron": job.get("cron"),
                "suggested_cron": job.get("cron"),  # 简化实现
                "reason": f"基于{goal}目标的调度优化"
            }
            suggestions.append(suggestion)
        
        return suggestions

    def _analyze_resource_trends(self, context: Dict) -> Dict:
        """分析资源趋势"""
        return {
            "cpu_trend": "stable",
            "memory_trend": "increasing",
            "worker_utilization": "moderate",
            "predicted_bottlenecks": []
        }

    def _extract_actionable_items(self, ai_text: str) -> List[str]:
        """从AI文本中提取可执行的建议"""
        # 简化实现，实际可以使用NLP技术提取
        items = []
        lines = ai_text.split('\n')
        for line in lines:
            if any(keyword in line for keyword in ['建议', '应该', '可以', '需要']):
                items.append(line.strip())
        return items[:5]  # 返回前5个建议

    def _generate_implementation_steps(self, schedule_text: str) -> List[str]:
        """生成实施步骤"""
        return [
            "1. 备份当前调度配置",
            "2. 在测试环境验证新调度方案",
            "3. 逐步应用新的调度设置",
            "4. 监控系统性能变化",
            "5. 根据运行结果进行微调"
        ]

    async def handle_tool_call(self, tool_name: str, arguments: Dict) -> Dict:
        """处理工具调用"""
        if tool_name not in self.tools:
            return {
                "success": False,
                "error": f"Unknown tool: {tool_name}"
            }
        
        try:
            result = await self.tools[tool_name](arguments)
            return result
        except Exception as e:
            self.logger.error(f"Tool execution failed: {e}")
            return {
                "success": False,
                "error": str(e)
            }

    def get_tool_definitions(self) -> List[Dict]:
        """获取工具定义"""
        return [
            {
                "name": "list_jobs",
                "description": "列出系统中的所有任务",
                "parameters": {
                    "department_id": {"type": "string", "description": "部门ID（可选）"},
                    "status": {"type": "string", "description": "任务状态（可选）"},
                    "limit": {"type": "integer", "description": "返回数量限制（默认10）"}
                }
            },
            {
                "name": "analyze_job_performance",
                "description": "分析任务执行性能",
                "parameters": {
                    "job_id": {"type": "string", "description": "任务ID（必需）"},
                    "days": {"type": "integer", "description": "分析天数（默认7天）"},
                    "metric": {"type": "string", "description": "分析指标"}
                }
            },
            {
                "name": "optimize_schedule",
                "description": "优化任务调度时间",
                "parameters": {
                    "job_ids": {"type": "array", "description": "任务ID列表"},
                    "goal": {"type": "string", "description": "优化目标"}
                }
            },
            {
                "name": "predict_resource_usage",
                "description": "预测系统资源使用情况",
                "parameters": {
                    "hours": {"type": "integer", "description": "预测小时数"},
                    "granularity": {"type": "string", "description": "时间粒度"}
                }
            },
            {
                "name": "get_recommendations",
                "description": "获取任务优化建议",
                "parameters": {
                    "job_id": {"type": "string", "description": "任务ID（可选）"},
                    "type": {"type": "string", "description": "建议类型"}
                }
            },
            {
                "name": "create_smart_schedule",
                "description": "创建智能调度方案",
                "parameters": {
                    "requirements": {"type": "object", "description": "调度需求"},
                    "constraints": {"type": "object", "description": "约束条件"}
                }
            }
        ]

# Main execution
async def main():
    server = GoJobMCPServer()
    
    while True:
        try:
            line = input()
            if not line:
                break
                
            request = json.loads(line)
            
            if request.get("method") == "tools/list":
                response = {
                    "tools": server.get_tool_definitions()
                }
            elif request.get("method") == "tools/call":
                tool_name = request.get("params", {}).get("name")
                arguments = request.get("params", {}).get("arguments", {})
                response = await server.handle_tool_call(tool_name, arguments)
            else:
                response = {
                    "error": f"Unknown method: {request.get('method')}"
                }
            
            print(json.dumps(response, ensure_ascii=False))
            sys.stdout.flush()
            
        except EOFError:
            break
        except Exception as e:
            error_response = {
                "error": str(e)
            }
            print(json.dumps(error_response))
            sys.stdout.flush()

if __name__ == "__main__":
    asyncio.run(main())
