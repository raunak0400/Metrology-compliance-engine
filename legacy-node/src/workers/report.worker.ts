/**
 * Report worker — consumes the `report-generation` queue. Collects the
 * inspection data, renders the report, uploads it to object storage, and marks
 * the Report row READY (or FAILED). The API then serves a signed download URL.
 */
import { Worker } from 'bullmq';
import { bullConnection } from '../config/redis';
import { logger } from '../config/logger';
import { buildObjectKey, putObject } from '../config/storage';
import { sha256 } from '../common/utils/crypto';
import { QUEUE_NAMES, type ReportJobData } from '../config/queues';
import { collectReportData, markGenerating, markReady, markFailed } from '../modules/reports/reports.service';
import { renderReport } from '../modules/reports/report.renderer';

export function startReportWorker(): Worker<ReportJobData> {
  const worker = new Worker<ReportJobData>(
    QUEUE_NAMES.REPORT,
    async (job) => {
      const { inspectionId, reportId, type } = job.data;
      logger.info({ jobId: job.id, reportId, inspectionId }, 'Generating report');

      try {
        await markGenerating(reportId);
        const data = await collectReportData(inspectionId);
        const rendered = renderReport(data as unknown as Record<string, unknown>, type);

        const key = buildObjectKey(
          `inspections/${inspectionId}/reports`,
          `report-${reportId}.${rendered.extension}`,
        );
        await putObject({ key, body: rendered.buffer, contentType: rendered.contentType });

        await markReady(reportId, {
          storageKey: key,
          fileName: `report-${reportId}.${rendered.extension}`,
          sizeBytes: rendered.buffer.length,
          checksum: sha256(rendered.buffer),
        });
        logger.info({ jobId: job.id, reportId, key }, 'Report ready');
      } catch (err) {
        await markFailed(reportId, (err as Error).message);
        throw err;
      }
    },
    { connection: bullConnection, concurrency: 2 },
  );

  worker.on('failed', (job, err) => logger.error({ jobId: job?.id, err }, 'Report job failed'));
  return worker;
}
